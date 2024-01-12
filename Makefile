toc:
	go run tools/toc_generator/main.go

teardown:
	- pkill -9 cockroach
	- docker ps -aq | xargs docker rm -f
	- docker rmi $(docker images | grep 'localhost:9090/app')
	- docker volume rm pg_primary pg_replica before_eu_db before_jp_db before_us_db bigquery
	- k3d cluster delete local
	- rm -rf inflight_trace_dump
	- rm -rf **/pgdata
	- rm -rf **/pg_archive
	- rm -Rf ~/.cassandra
