GO111MODULE=on

.PHONY: all test test-short build

build:
	go build ./...

test-short: build
	go test ./... -v -covermode=count -coverprofile=coverage.out -short

test: build
	go test ./... -covermode=count -coverprofile=coverage.out

push: build
	git push -f dokku main

docker-build:
	sudo docker build --build-arg ENV=prod -t dokku-home .
	
docker-install: docker-build
	sudo docker run --name my-dokku-home -d -p 0.0.0.0:8080:8080 dokku-home

docker-prompt:
	sudo docker exec -it my-dokku-home /bin/sh

docker-del:
	sudo docker kill my-dokku-home
	sudo docker stop my-dokku-home
	sudo docker rm my-dokku-home

docker-rm: docker-del
	sudo docker image rm dokku-home

