package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/etng/SysProbe/probe"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

func main() {
	log.Printf("SysProbe v%s", probe.VERSION)
	var (
		pushURL    string
		limit      int
		interval   int
		httpListen string
		debugMode  bool
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
	if limit > 1 {
		log.Printf("config check_times %s", limit)
	} else {
		log.Printf("config check only one time")
	}
	if len(pushURL) > 0 {
		log.Printf("config push_url %s", pushURL)
		RepeatCall(time.Duration(interval)*time.Second, limit, func() {
			notifyProbe(pushURL)
		})
	}

	if len(httpListen) > 0 {
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
		RepeatCall(time.Duration(interval)*time.Second, limit, showProbe)
	}
	fmt.Println("done")
}

func notifyProbe(url string) {
	if resp, err := http.Post(url, "application/json", strings.NewReader(getProbe())); err == nil {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			log.Printf("notify response for %s is : %s", url, string(body))
		} else {
			log.Printf("fail to read response %s %v %v:", url, resp, err)
		}
	} else {
		log.Printf("fail to notify %s %v %v:", url, resp, err)
	}

}
func getProbe() string {
	result := probe.Probe()
	b, _ := json.MarshalIndent(result, "", " ")
	return string(b)
}
func showProbe() {
	log.Println(getProbe())
}

// RepeatCall 按间隔执行指定的函数并在相关次数后停止,首次执行不延时
func RepeatCall(interval time.Duration, limit int, handler func()) {
	handler()
	if limit == 1 {
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
			go func() {
				handler()
				atomic.AddInt32(times, 1)
				if limit > 0 && atomic.LoadInt32(times) >= int32(limit) {
					stopChan <- 1
				}
			}()
		}
	}
}
