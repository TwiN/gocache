docker-build:
	docker build --tag=gocache-server .

docker-run:
	docker run -p 6666:6379 --name gocache-server -d gocache-server