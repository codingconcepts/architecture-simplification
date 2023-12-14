toc:
	go run tools/toc_generator/main.go

teardown:
	- pkill -9 cockroach
	- docker ps -aq | xargs docker rm -f
	- docker volume rm pg-data
	- rm -rf inflight_trace_dump **/inflight_trace_dump **/pgdata
	- k3d cluster delete local