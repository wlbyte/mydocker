package subsystems

import (
	"fmt"
	"os"
	"path"
	"strconv"
)

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}

func (s *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.Cpus == "" {
		return nil
	}
	errFormat := "cpuSubsystem.Set: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	cpus, err := strconv.ParseFloat(res.Cpus, 64)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.WriteFile(path.Join(subsysPath, "cpu.cfs_quota_us"), []byte(strconv.Itoa(int(100000*cpus))), 0644); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int, res *ResourceConfig) error {
	if res.Cpus == "" {
		return nil
	}
	errFormat := "cpuSubsystem.Apply: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.WriteFile(path.Join(subsysPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	errFormat := "cpuSubsystem.Remove: %w"
	subsysPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.RemoveAll(subsysPath); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}
