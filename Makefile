build-docker:
	docker build -t crossbar-image .

run-docker:
	docker-compose -f docker-compose.yml run --rm crossbar
