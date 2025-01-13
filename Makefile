build-docker:
	docker build -t crossbar-image .

run-docker:
	docker-compose -f docker-compose.yml run --rm crossbar

build-docker-xconn:
	docker build -f Dockerfile.xconn -t xconn-image .

run-docker-xconn:
	docker-compose -f docker-compose.yml run --rm xconn
