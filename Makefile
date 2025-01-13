build-docker-crossbar:
	docker build -t crossbar-image -f Dockerfile.crossbar .

run-docker-crossbar:
	docker compose -f docker-compose.yml run --rm crossbar

build-docker-xconn:
	docker build -f Dockerfile.xconn -t xconn-image .

run-docker-xconn:
	docker compose -f docker-compose.yml run --rm xconn
