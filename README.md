# SysProbe

Just for you to inspect your system infomation and push to your data center or fetch it from your data center.

## Usage

### one time
```bash
./probe
```

### limit time
```bash
./probe -limit 100
```
### limit time with interval 
```bash
./probe -limit 100 -interval 5
```
### push to url 

```bash
./probe -push_url http://example.com/api/v1/host
```
the url must accept json data with http post method
you can also combine the `-limit <number>` and `-interval <nubner>` parameter
### listen with http address

```bash
./probe -http_listen :1234
```
you can get the result with your browser or cli

```bash
curl localhost:1234 | jq
```
## Data Example

```json
{
 "bootTime": 1572692666,
 "diskFree": 82003193856,
 "diskFreeI": 51903882,
 "diskFreePercent": 76.3796304717792,
 "diskFreePercentI": 98.99976516494301,
 "diskTotal": 107362648064,
 "diskTotalI": 52428288,
 "diskUsed": 25359454208,
 "diskUsedI": 524406,
 "hostId": "b30d0f21-10ac-3807-b210-c19ede3ce88f",
 "hostName": "ddcd-y10n",
 "ifs": {
  "52:54:00:be:92:a3": "192.168.99.202/24,fe80::5054:ff:febe:92a3/64"
 },
 "load1": 0,
 "load15": 0.12,
 "load5": 0.08,
 "memFree": 1019416576,
 "memFreeRatio": 30.492622301594196,
 "memTotal": 8201773056,
 "memUsed": 5700837376,
 "osType": "centos",
 "osVersion": "7.5.1804",
 "procsBlocked": 0,
 "procsRunning": 1,
 "procsTotal": 958,
 "time": "2019-11-28T11:39:11+08:00",
 "ts": 1574912351,
 "upTime": 2219685
}
```
