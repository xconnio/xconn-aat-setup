build-docker-crossbar:
	docker build -t crossbar-image -f Dockerfile.crossbar .

run-docker-crossbar:
	docker compose -f docker-compose.yml run --rm crossbar

build-docker-xconn:
	docker build -f Dockerfile.xconn -t xconn-image .

run-docker-xconn:
	docker compose -f docker-compose.yml run --rm xconn

run-wick-commands:
	wick call io.xconn.backend.add2 2 4 --url "$(url)"
	wick call io.xconn.backend.add2 2 4 --url "$(url)" --authid cryptosign-user --private-key 150085398329d255ad69e82bf47ced397bcec5b8fbeecd28a80edbbd85b49081
	wick call io.xconn.backend.add2 2 4 --url "$(url)" --ticket ticket-pass --authid ticket-user
	wick call io.xconn.backend.add2 2 4 --url "$(url)" --authid wamp-cra-user --secret cra-secret
	wick call io.xconn.backend.add2 2 4 --authid wamp-cra-salt-user --secret cra-salt-secret
