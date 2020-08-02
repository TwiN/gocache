docker-build:
	docker build --tag=gocache-server .

docker-run:
	docker run -p 6666:6379 --name gocache-server -d -m=512m gocache-server

run:
	PORT=6666 go run examples/server/server.go

redis-benchmark:
	redis-benchmark -p 6666 -t set,get -n 10000000 -r 200000 -q -P 512 -c 512

memtier-benchmark:
	memtier_benchmark --port 6666 --hide-histogram --key-maximum 100000 --ratio 1:1 --expiry-range 1-100 --key-pattern R:R --randomize -n 100000