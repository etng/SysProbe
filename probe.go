package main

import (
	"fmt"
	"strings"
	"time"
	"flag"
	"io"
	"log"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
	"sync/atomic"

)

// VERSION 当前探测器版本
const VERSION = "1.0Beta"

func main() {
	log.Printf("SysProbe v%s", VERSION)
	var (
		pushURL  string
		limit int
		interval  int
		httpListen  string
		debugMode bool
	)
	flag.IntVar(&limit, "limit", 0, "limit for check")
	flag.IntVar(&interval, "interval", 0, "interval for check")
	flag.StringVar(&pushURL, "push_url", "", "if you want to push result to remote, set it")
	flag.StringVar(&httpListen, "http_listen", "", "if you want to serve the result by http api, configure it as :1234")
	flag.BoolVar(&debugMode, "debug", false, "use true to only open mogu_center related components")
	flag.Parse()
	if interval == 0 {
		// 如果间隔为0则表示只执行一次
		limit = 1
	} else {
		log.Printf("config push_interval %s second(s)", interval)
	}
	if limit > 1{
		log.Printf("config check_times %s", limit)
	} else{
		log.Printf("config check only one time")
	}
	if len(pushURL) >0{
		log.Printf("config push_url %s", pushURL)
		RepeatCall(time.Duration(interval) * time.Second, limit, func(){
			notifyProbe(pushURL)
		})
	}

	if len(httpListen)>0{
		log.Printf("config httpListen %s", httpListen)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			log.Printf("status check from %s request received", r.RemoteAddr)
			defer log.Printf("status check from %s request handled", r.RemoteAddr)
			io.WriteString(w, getProbe())	
		})
		log.Fatal(http.ListenAndServe(httpListen, nil))
	} else {
		RepeatCall(time.Duration(interval) * time.Second, limit, showProbe)
	}
	fmt.Println("done")
}

func notifyProbe(url string){
	if resp, err := http.Post(url, "application/json", strings.NewReader(getProbe())); err==nil{
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err==nil{
			log.Printf("notify response for %s is : %s", url, string(body))
		}else{
			log.Printf("fail to read response %s %v %v:", url, resp, err)	
		}
	} else {
		log.Printf("fail to notify %s %v %v:", url, resp, err)
	}

}
func getProbe() string{
	result := Probe()
	b, _ := json.MarshalIndent(result, "", " ")
	return string(b)
}
func showProbe(){
	log.Println(getProbe())
}

// ValidateEthName 判断网卡是否是网桥或者docker虚拟的或者是回环地址
func ValidateEthName(s string)(bool){
	prefixes := []string{
		"br-",
		"lo",
		"docker",
		"veth",
	}
	for _, prefix := range prefixes{
		if strings.HasPrefix(s, prefix) {
			return false
		}
	}
	return true
}

// Probe 提取当前系统基本信息
func Probe()(map[string]interface{}){
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
	if _ifs, e := net.Interfaces();e==nil{
		for _, _if := range _ifs{
			if ValidateEthName(_if.Name){
				ifs[_if.HardwareAddr] = strings.Join(func()([]string){ 
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

// RepeatCall 按间隔执行指定的函数并在相关次数后停止,首次执行不延时
func RepeatCall(interval time.Duration, limit int, handler func()){
	handler()
	if limit == 1{
		return
	}
	ticker := time.NewTicker(interval)
	_times := int32(1)
	var times *int32 = &_times
	defer ticker.Stop()
	// 持续操作
	stopChan := make(chan int, 0)
	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			go func(){
				handler()
				atomic.AddInt32(times, 1)
				if limit> 0 && atomic.LoadInt32(times) >= int32(limit) {
					stopChan<-1
				}
			}()
		}
	}
}