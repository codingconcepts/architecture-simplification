toc:
	go run tools/toc_generator/main.go

teardown:
	- pkill -9 cockroach
	- docker ps -aq | xargs docker rm -f
	- docker volume rm pg_primary pg_replica
	- rm -rf inflight_trace_dump
	- rm -rf **/inflight_trace_dump
	- rm -if **/pgdata **/pg_archive
	- k3d cluster delete local