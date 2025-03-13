package subsystems

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

type MemorySubSystem struct {
}

func (s *MemorySubSystem) Name() string {
	return "memory"
}

func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.MemoryLimit == "" {
		return nil
	}
	errFormat := "memorySubSystem.Set: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.WriteFile(path.Join(subsysPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func (s *MemorySubSystem) Apply(cgroupPath string, pid int, res *ResourceConfig) error {
	if res.MemoryLimit == "" {
		return nil
	}
	errFormat := "memorySubSystem.Apply: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.WriteFile(path.Join(subsysPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func (s *MemorySubSystem) Remove(cgroupPath string) error {
	errFormat := "memorySubSystem.Apply: %s: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf(errFormat, "getCgroupPath", err)
	}
	if err := os.RemoveAll(subsysPath); err != nil {
		return fmt.Errorf(errFormat, "os.RemoveAll", err)
	}
	return nil
}
