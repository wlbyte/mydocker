package subsystems

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.CpuSet == "" {
		return nil
	}
	errFormat := "cpusetSubSystem.Set error: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if res.CpuSet == "" {
		return nil
	}
	if err := os.WriteFile(path.Join(subsysPath, "cpuset.cpus"), []byte(res.CpuSet), 0644); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int, res *ResourceConfig) error {
	if res.CpuSet == "" {
		return nil
	}
	errFormat := "cpusetSubSystem.Apply error: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.WriteFile(path.Join(subsysPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {

	errFormat := "cpusetSubSystem.Remove error: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.RemoveAll(subsysPath); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}
