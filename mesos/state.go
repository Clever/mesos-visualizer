package mesos

type State struct {
	ActivatedSlaves int64   `json:"activated_slaves"`
	BuildDate       string  `json:"build_date"`
	BuildTime       float64 `json:"build_time"`
	BuildUser       string  `json:"build_user"`
	// CompletedFrameworks    []Something       `json:"completed_frameworks"`
	DeactivatedSlaves int64             `json:"deactivated_slaves"`
	ElectedTime       float64           `json:"elected_time"`
	FailedTasks       int64             `json:"failed_tasks"`
	FinishedTasks     int64             `json:"finished_tasks"`
	Flags             map[string]string `json:"flags"`
	Frameworks        []Framework       `json:"frameworks"`
	GitSHA            string            `json:"git_sha"`
	GitTag            string            `json:"git_tag"`
	Hostname          string            `json:"hostname"`
	ID                string            `json:"id"`
	KilledTasks       int64             `json:"killed_tasks"`
	Leader            string            `json:"leader"`
	LogDir            string            `json:"log_dir"`
	LostTasks         int64             `json:"lost_tasks"`
	// OrphanTasks            []Something       `json:"orphan_tasks"`
	PID          string  `json:"pid"`
	Slaves       []Slave `json:"slaves"`
	StagedTasks  int64   `json:"staged_tasks"`
	StartTime    float64 `json:"start_time"`
	StartedTasks int64   `json:"started_tasks"`
	// UnregisteredFrameworks []Something       `json:"unregistered_frameworks"`
	Version string `json:"version"`
}

type Slave struct {
	Active         bool              `json:"active"`
	Attributes     map[string]string `json:"attributes"`
	Hostname       string            `json:"hostname"`
	ID             string            `json:"id"`
	PID            string            `json:"pid"`
	RegisteredTime float64           `json:"registered_time"`
	Resources      Resources         `json:"resources"`
}

type Resources struct {
	CPUs  float64 `json:"cpus"`
	Disk  int64   `json:"disk"`
	Mem   int64   `json:"mem"`
	Ports string  `json:"ports"`
}

type Framework struct {
	Active           bool        `json:"active"`
	Checkpoint       bool        `json:"checkpoint"`
	CompletedTasks   []Task      `json:"checkpoint"`
	FailoverTimeout  int64       `json:"failover_timeout"`
	Hostname         string      `json:"hostname"`
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	OfferedResources Resources   `json:"offered_resources"`
	Offers           []Resources `json:"offers"`
	RegisteredTime   float64     `json:"registered_time"`
	ReregisteredTime float64     `json:"reregistered_time"`
	Resources        Resources   `json:"resources"`
	Role             string      `json:"role"`
	Tasks            []Task      `json:"tasks"`
	UnregisteredTime float64     `json:"unregistered_time"`
	UsedResources    Resources   `json:"used_resources"`
	User             string      `json:"user"`
	WebUIURL         string      `json:"webui_url"`
}

type Task struct {
	ExecutorID  string    `json:"executor_id"`
	FrameworkID string    `json:"framework_id"`
	ID          string    `json:"id"`
	Labels      []Label   `json:"labels"`
	Name        string    `json:"name"`
	Resources   Resources `json:"resources"`
	SlaveID     string    `json:"slave_id"`
	State       string    `json:"state"`
	Statuses    []Status  `json:"statuses"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Status struct {
	State     string  `json:"state"`
	Timestamp float64 `json:"timestamp"`
}
