// +build darwin
// +build !cgo

package mem

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/internal/common"
)

// Runs vm_stat and returns Free and inactive pages
func getVMStat(vms *VirtualMemoryStat) error {
	out, err := exec.Command("vm_stat").Output()
	if err != nil {
		return err
	}
	return parseVMStat(string(out), vms)
}

func parseVMStat(out string, vms *VirtualMemoryStat) error {
	var err error

	lines := strings.Split(out, "\n")
	pagesize := uint64(syscall.Getpagesize())
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.Trim(fields[1], " .")
		switch key {
		case "Pages free":
			free, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Free = free * pagesize
		case "Pages inactive":
			inactive, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Inactive = inactive * pagesize
		case "Pages active":
			active, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Active = active * pagesize
		case "Pages wired down":
			wired, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Wired = wired * pagesize
		}
	}
	return err
}

// VirtualMemory returns VirtualmemoryStat.
func VirtualMemory() (*VirtualMemoryStat, error) {
	ret := &VirtualMemoryStat{}

	t, err := common.DoSysctrl("hw.memsize")
	if err != nil {
		return nil, err
	}
	total, err := strconv.ParseUint(t[0], 10, 64)
	if err != nil {
		return nil, err
	}
	err = getVMStat(ret)
	if err != nil {
		return nil, err
	}

	ret.Available = ret.Free + ret.Inactive
	ret.Total = total

	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Total-ret.Available) / float64(ret.Total) * 100.0

	return ret, nil
}