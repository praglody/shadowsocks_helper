release : ss_server ss_local

ss_server : server/server.go
	go build -o ss_server server/server.go

ss_local : local/local.go
	go build -o ss_local local/local.go

clean :
	rm -f ss_server ss_local

