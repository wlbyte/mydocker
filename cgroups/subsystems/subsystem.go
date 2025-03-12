package subsystems

var (
	SubsystemsIns = []Subsystem{
		&CpuSubSystem{},
		&MemorySubSystem{},
		&CpusetSubSystem{},
	}
)

type ResourceConfig struct {
	MemoryLimit string
	Cpus        string
	CpuSet      string
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int, res *ResourceConfig) error
	Remove(path string) error
}
