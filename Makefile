DEV_LISTEN=:1234
DEV_PUSH=http://localhost:80/anything
help:
	cat Makefile
build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o bin/probe_linux_amd64 probe.go && upx bin/probe_linux_amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s" -o bin/probe_mac_amd64 probe.go && upx bin/probe_mac_amd64
	GOOS=windows GOARCH=amd64 go build -ldflags "-w -s" -o bin/probe_win_amd64.exe probe.go && upx bin/probe_win_amd64.exe
	ln -f -s bin/$(go env GOOS)_$(go env GOHOSTARCH) bin/probe
run:
	./bin/probe -http_listen ${DEV_LISTEN} -push_url ${DEV_PUSH}
client:
	curl http://localhost${DEV_LISTEN}
dev:
	go run probe.go -http_listen ${DEV_LISTEN} -push_url ${DEV_PUSH}
proxy:
	export GOPROXY=https://goproxy.cn,direct


