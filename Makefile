SHELL=/usr/bin/env bash

docker-build:
	sudo docker build . -t mes3
docker-run:
	sudo docker run --restart=always -d --name filData --network blockchain-browser_frontend -p 9000:9000 mes3
