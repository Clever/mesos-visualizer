package mesos

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type ResourceNode struct {
	Name        string         `json:"name"`
	Children    []ResourceNode `json:"children,omitempty"`
	Memory      int64          `json:"memory,omitempty"`
	CPU         float64        `json:"cpu,omitempty"`
	MemoryTotal int64          `json:"memory_total,omitempty"`
	CPUTotal    float64        `json:"cpu_total,omitempty"`
}

type Client struct {
	client *http.Client
	host   string
}

func NewClient(mesosHost string) *Client {
	return &Client{
		client: &http.Client{},
		host:   mesosHost,
	}
}

func (c *Client) GetState() (State, error) {
	stateURL := url.URL{Scheme: "http", Host: c.host, Path: "/state.json"}
	var decodedResponse State
	err := c.Get(stateURL, &decodedResponse)
	if err != nil {
		return decodedResponse, err
	}
	// strip "master@" from beginning of leader
	leader := decodedResponse.Leader[7:]

	leaderURL := url.URL{Scheme: "http", Host: leader, Path: "/state.json"}
	err = c.Get(leaderURL, &decodedResponse)
	return decodedResponse, err
}

func (c *Client) Get(endpoint url.URL, response interface{}) error {
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	jsonDecoder := json.NewDecoder(resp.Body)
	if err := jsonDecoder.Decode(&response); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetResourceGraph() (ResourceNode, error) {
	state, err := c.GetState()
	if err != nil {
		return ResourceNode{}, err
	}

	memTotal := int64(0)
	cpuTotal := 0.0

	slaveToTaskMap := buildSlaveToTaskMap(state)

	slaveNodes := []ResourceNode{}
	for _, slave := range state.Slaves {
		slaveCPU := 0.0
		slaveMem := int64(0)

		cpuTotal += slave.Resources.CPUs
		memTotal += slave.Resources.Mem

		slaveNode := ResourceNode{
			Name:        slave.Hostname,
			Children:    []ResourceNode{},
			CPUTotal:    slave.Resources.CPUs,
			MemoryTotal: slave.Resources.Mem,
		}

		tasks, ok := slaveToTaskMap[slave.ID]
		if ok {
			for _, task := range tasks {
				taskNode := ResourceNode{
					Name:   task.Name,
					Memory: task.Resources.Mem,
					CPU:    task.Resources.CPUs,
				}
				slaveNode.Children = append(slaveNode.Children, taskNode)
				slaveCPU += task.Resources.CPUs
				slaveMem += task.Resources.Mem
			}
		}

		slaveUnused := ResourceNode{
			Name:   "Unused",
			Memory: slave.Resources.Mem - slaveMem,
			CPU:    slave.Resources.CPUs - slaveCPU,
		}
		slaveNode.Memory = slaveMem
		slaveNode.CPU = slaveCPU
		slaveNode.Children = append(slaveNode.Children, slaveUnused)
		slaveNodes = append(slaveNodes, slaveNode)
	}

	root := ResourceNode{
		Name:        "Total",
		CPUTotal:    cpuTotal,
		MemoryTotal: memTotal,
		CPU:         state.Frameworks[0].Resources.CPUs,
		Memory:      state.Frameworks[0].Resources.Mem,
		Children:    slaveNodes,
	}

	return root, nil
}

func buildSlaveToTaskMap(state State) map[string][]Task {
	slaveToTaskMap := map[string][]Task{}
	framework := state.Frameworks[0]
	for _, task := range framework.Tasks {
		if task.State == "TASK_RUNNING" {
			if _, ok := slaveToTaskMap[task.SlaveID]; !ok {
				slaveToTaskMap[task.SlaveID] = []Task{}
			}
			slaveToTaskMap[task.SlaveID] = append(slaveToTaskMap[task.SlaveID], task)
		}
	}
	return slaveToTaskMap
}
