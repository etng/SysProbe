package probe

import (
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"strings"
	"time"
)

// VERSION 当前探测器版本
const VERSION = "1.0Beta"

// ValidateEthName 判断网卡是否是网桥或者docker虚拟的或者是回环地址
func ValidateEthName(s string) bool {
	prefixes := []string{
		"br-",
		"lo",
		"docker",
		"veth",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return false
		}
	}
	return true
}

// Probe 提取当前系统基本信息
func Probe() map[string]interface{} {
	result := map[string]interface{}{}

	avg, _ := load.Avg()
	now := time.Now()
	result["ts"] = now.Unix()
	result["time"] = now.Format(time.RFC3339)
	result["load1"] = avg.Load1
	result["load5"] = avg.Load5
	result["load15"] = avg.Load15
	misc, _ := load.Misc()
	result["procsTotal"] = misc.ProcsTotal
	result["procsRunning"] = misc.ProcsRunning
	result["procsBlocked"] = misc.ProcsBlocked

	vm, _ := mem.VirtualMemory()
	result["memTotal"] = vm.Total
	result["memUsed"] = vm.Used
	result["memFree"] = vm.Free
	result["memFreeRatio"] = 100.0 - vm.UsedPercent

	du, _ := disk.Usage("/")
	result["diskTotal"] = du.Total
	result["diskFree"] = du.Free
	result["diskUsed"] = du.Used
	result["diskFreePercent"] = 100.0 - du.UsedPercent

	result["diskTotalI"] = du.InodesTotal
	result["diskFreeI"] = du.InodesFree
	result["diskUsedI"] = du.InodesUsed
	result["diskFreePercentI"] = 100.0 - du.InodesUsedPercent

	info, _ := host.Info()
	result["hostName"] = info.Hostname
	result["hostId"] = info.HostID
	result["upTime"] = info.Uptime
	result["bootTime"] = info.BootTime
	result["osType"] = info.Platform
	result["osVersion"] = info.PlatformVersion

	ifs := map[string]interface{}{}
	if _ifs, e := net.Interfaces(); e == nil {
		for _, _if := range _ifs {
			if ValidateEthName(_if.Name) {
				ifs[_if.HardwareAddr] = strings.Join(func() []string {
					var addrs []string
					for _, _addr := range _if.Addrs {
						addrs = append(addrs, _addr.Addr)
					}
					return addrs
				}(), ",")
			}
		}
	}
	result["ifs"] = ifs
	return result
}
