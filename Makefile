toc:
	go run tools/toc_generator/main.go

teardown:
	- pkill -9 cockroach
	- docker ps -aq | xargs docker rm -f
	- rm -rf inflight_trace_dump
	- k3d cluster delete local