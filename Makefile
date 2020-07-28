docker-build:
	docker build --tag=gocache-server .

docker-run:
	docker run -p 6666:6379 --name gocache-server -d -m=512m gocache-server

run:
	PORT=6666 go run examples/server/server.go