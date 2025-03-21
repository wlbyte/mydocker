package subsystems

var (
	SubsystemsIns = []Subsystem{
		&CpuSubSystem{},
		&MemorySubSystem{},
		&CpusetSubSystem{},
	}
)

type ResourceConfig struct {
	MemoryLimit string `json:"memoryLimit"`
	Cpus        string `json:"cpus"`
	CpuSet      string `json:"cpuset"`
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int, res *ResourceConfig) error
	Remove(path string) error
}
