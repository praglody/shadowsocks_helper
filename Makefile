release : ss_server ss_local

ss_server : server/server.go
	go build -o ss_server server/server.go

ss_local : local/local.go
	go build -o ss_local local/local.go

clean :
	rm -f ss_server ss_local

linux_release :
	rm -f linux_ss_server linux_ss_local
	GOOS=linux GOARCH=amd64 go build -o linux_ss_server server/server.go
	GOOS=linux GOARCH=amd64 go build -o linux_ss_local local/local.go

