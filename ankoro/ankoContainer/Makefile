MODULE=ankoro
ANKO=anko
GOOS=linux

install_container: build build_container register_container

build: 
	GOOS=$(GOOS) go build -o $(ANKO) $(GOPATH)/src/github.com/mattn/anko/anko.go

build_container: Dockerfile
	docker build -t $(MODULE) --build-arg ANKO=$(ANKO) .

register_container:
	docker tag $(MODULE) localhost:5000/$(MODULE)
	docker push localhost:5000/$(MODULE)
	-docker rmi localhost:5000/$(MODULE) -f
	docker pull localhost:5000/$(MODULE)
	-docker rmi $$(docker images -a --filter "dangling=true" -q) -f

start:
	docker run -t -i -d $(MODULE) 
	#docker run --rm -t -i $(MODULE) 
	#docker run --rm -t -i $(MODULE) /bin/bash
stop:
	-docker stop $$( docker ps -a | grep $(MODULE) | awk '{print $$1}' )
	docker rm $$( docker ps -a | grep $(MODULE) | awk '{print $$1}' )
attach:
	docker attach $$( docker ps -a | grep $(MODULE) | awk '{print $$1}' )
attach_stdout:
	docker attach $$( docker ps -a | grep $(MODULE) | awk '{print $$1}' )  --no-stdin
