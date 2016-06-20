package ecs

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ResourceNode struct {
	Name        string         `json:"name"`
	Children    []ResourceNode `json:"children,omitempty"`
	Memory      float64        `json:"memory,omitempty"`
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

	for _, containerInstanceARN := range containerInstanceARNs {
		describeContainerInstancesInput := &ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(clusterName),
			ContainerInstances: []*string{containerInstanceARN},
		}
		resp, err := c.client.DescribeContainerInstances(describeContainerInstancesInput)
		if err != nil {
			return ResourceNode{}, err
		}

		if len(resp.ContainerInstances) != 1 {
			return ResourceNode{}, errors.New("wrong number of container instances")
		}

		containerInstance := resp.ContainerInstances[0]
		slaveNode := ResourceNode{
			Name:     *containerInstance.Ec2InstanceId,
			Children: []ResourceNode{},
		}

		var usedCPU float64
		var usedMem float64
		var remainingCPU float64
		var remainingMem float64

		for _, resource := range containerInstance.RegisteredResources {
			switch *resource.Name {
			case "CPU":
				usedCPU += float64(*resource.IntegerValue)
			case "MEMORY":
				usedMem += float64(*resource.IntegerValue)
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

		slaveNode.CPUTotal = usedCPU
		slaveNode.MemoryTotal = usedMem
		slaveNode.CPU = usedCPU - remainingCPU
		slaveNode.Memory = usedMem - remainingMem

		cpuTotal += usedCPU
		memTotal += usedMem
		cpuUsedTotal += usedCPU - remainingCPU
		memUsedTotal += usedMem - remainingMem

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

		describeTasksInput := &ecs.DescribeTasksInput{
			Cluster: aws.String(clusterName),
			Tasks:   taskARNs,
		}

		tresp, err := c.client.DescribeTasks(describeTasksInput)
		if err != nil {
			return ResourceNode{}, err
		}

		for _, task := range tresp.Tasks {
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

			taskNode := ResourceNode{
				Name:   *cd.Name,
				Memory: float64(*cd.Memory),
				CPU:    float64(*cd.Cpu),
			}
			slaveNode.Children = append(slaveNode.Children, taskNode)
		}
		slaveUnused := ResourceNode{
			Name:   "Unused",
			Memory: remainingMem,
			CPU:    remainingCPU,
		}
		slaveNode.Children = append(slaveNode.Children, slaveUnused)
		slaveNodes = append(slaveNodes, slaveNode)

	}

	root := ResourceNode{
		Name:        "Total",
		CPUTotal:    cpuTotal,
		MemoryTotal: memTotal,
		CPU:         cpuUsedTotal,
		Memory:      memUsedTotal,
		Children:    slaveNodes,
	}
	return root, nil
}
