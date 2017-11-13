package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ResourceNode struct {
	Name        string         `json:"name"`
	Children    []ResourceNode `json:"children,omitempty"`
	SoftMemory  float64        `json:"soft_memory,omitempty"`
	MaxMemory   float64        `json:"max_memory,omitempty"`
	CPU         float64        `json:"cpu,omitempty"`
	MemoryTotal float64        `json:"memory_total,omitempty"`
	CPUTotal    float64        `json:"cpu_total,omitempty"`
}

type Client struct {
	client  *ecs.ECS
	cluster string
}

var taskDefinitionCache = map[string]*ecs.TaskDefinition{}

func NewClient(cluster string, accessKeyID string, secretAccessKey string) *Client {
	awsConfig := &aws.Config{
		Region:      aws.String("us-west-1"),
		MaxRetries:  aws.Int(10),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	}
	svc := ecs.New(session.New(), awsConfig)
	return &Client{
		client:  svc,
		cluster: cluster,
	}
}

func (c *Client) GetResourceGraph() (ResourceNode, error) {
	clusterName := c.cluster

	memTotal := 0.0
	cpuTotal := 0.0
	memUsedTotal := 0.0
	maxMemTotal := 0.0
	cpuUsedTotal := 0.0

	containerInstanceARNs := []*string{}
	listContainerInstancesInput := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	}
	if err := c.client.ListContainerInstancesPages(listContainerInstancesInput,
		func(p *ecs.ListContainerInstancesOutput, lastPage bool) bool {
			containerInstanceARNs = append(containerInstanceARNs, p.ContainerInstanceArns...)
			return !lastPage
		},
	); err != nil {
		return ResourceNode{}, err
	}

	slaveNodes := []ResourceNode{}

	describeContainerInstancesInput := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(clusterName),
		ContainerInstances: containerInstanceARNs,
	}
	dresp, err := c.client.DescribeContainerInstances(describeContainerInstancesInput)
	if err != nil {
		return ResourceNode{}, err
	}

	for _, containerInstance := range dresp.ContainerInstances {
		slaveNode := ResourceNode{
			Name:     *containerInstance.Ec2InstanceId,
			Children: []ResourceNode{},
		}

		var usableCPU float64
		var usableMem float64
		var remainingCPU float64
		var remainingMem float64
		var maxPossibleMem float64

		for _, resource := range containerInstance.RegisteredResources {
			switch *resource.Name {
			case "CPU":
				usableCPU += float64(*resource.IntegerValue)
			case "MEMORY":
				usableMem += float64(*resource.IntegerValue)
			}
		}

		for _, resource := range containerInstance.RemainingResources {
			switch *resource.Name {
			case "CPU":
				remainingCPU += float64(*resource.IntegerValue)
			case "MEMORY":
				remainingMem += float64(*resource.IntegerValue)
			}
		}

		slaveNode.CPUTotal = usableCPU
		slaveNode.MemoryTotal = usableMem
		slaveNode.CPU = usableCPU - remainingCPU
		slaveNode.SoftMemory = usableMem - remainingMem

		cpuTotal += usableCPU
		memTotal += usableMem
		cpuUsedTotal += usableCPU - remainingCPU
		memUsedTotal += usableMem - remainingMem

		listTasksInput := &ecs.ListTasksInput{
			Cluster:           aws.String(clusterName),
			ContainerInstance: containerInstance.ContainerInstanceArn,
		}
		taskARNs := []*string{}
		if err := c.client.ListTasksPages(listTasksInput, func(p *ecs.ListTasksOutput, lastPage bool) bool {
			taskARNs = append(taskARNs, p.TaskArns...)
			return !lastPage
		}); err != nil {
			return ResourceNode{}, err
		}

		idChunks := [][]*string{}
		for len(taskARNs) > 0 {
			end := 100
			if len(taskARNs) < 100 {
				end = len(taskARNs)
			}

			idChunks = append(idChunks, taskARNs[:end])
			taskARNs = taskARNs[end:]
		}

		tasks := []*ecs.Task{}
		for _, idChunk := range idChunks {
			describeTasksInput := &ecs.DescribeTasksInput{
				Cluster: aws.String(clusterName),
				Tasks:   idChunk,
			}
			tresp, err := c.client.DescribeTasks(describeTasksInput)
			if err != nil {
				return ResourceNode{}, err
			}
			tasks = append(tasks, tresp.Tasks...)
		}

		for _, task := range tasks {
			td, ok := taskDefinitionCache[*task.TaskDefinitionArn]
			if !ok {
				describeTaskDefinitionInput := &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: task.TaskDefinitionArn,
				}
				resp, err := c.client.DescribeTaskDefinition(describeTaskDefinitionInput)
				if err != nil {
					return ResourceNode{}, err
				}
				td = resp.TaskDefinition
				taskDefinitionCache[*task.TaskDefinitionArn] = td
			}

			cd := td.ContainerDefinitions[0]

			var soft_mem, max_mem float64
			max_mem = float64(*cd.Memory)
			if cd.MemoryReservation != nil {
				soft_mem = float64(*cd.MemoryReservation)
			}

			taskNode := ResourceNode{
				Name:       *cd.Name,
				MaxMemory:  max_mem,
				SoftMemory: soft_mem,
				CPU:        float64(*cd.Cpu),
			}
			slaveNode.Children = append(slaveNode.Children, taskNode)
			maxMemTotal += max_mem
			maxPossibleMem += max_mem
		}
		slaveUnused := ResourceNode{
			Name:       "Unused",
			SoftMemory: remainingMem,
			MaxMemory:  remainingMem,
			CPU:        remainingCPU,
		}
		slaveNode.MaxMemory = maxPossibleMem
		slaveNode.Children = append(slaveNode.Children, slaveUnused)
		slaveNodes = append(slaveNodes, slaveNode)

	}

	root := ResourceNode{
		Name:        "Total",
		CPUTotal:    cpuTotal,
		MemoryTotal: memTotal,
		CPU:         cpuUsedTotal,
		MaxMemory:   maxMemTotal,
		SoftMemory:  memUsedTotal,
		Children:    slaveNodes,
	}
	return root, nil
}
