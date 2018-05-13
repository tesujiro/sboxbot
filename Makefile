MODULE=sbox
GOOS=linux

all: build container deploy_container

deploy_container:
	for p in `kubectl get pod | grep $(MODULE) | awk '{print $$1}'`;do\
		kubectl delete pod $$p;\
	done

build: $(MODULE)

container: build_container register

$(MODULE): sbox.go twitter.go docker.go
	GOOS=$(GOOS) go build -o $(MODULE) sbox.go twitter.go docker.go
	
build_container: Dockerfile $(MODULE)
	docker build -t $(MODULE) .

register:
	docker tag $(MODULE) localhost:5000/$(MODULE)
	docker push localhost:5000/$(MODULE)
	docker rmi localhost:5000/$(MODULE)
	docker pull localhost:5000/$(MODULE)

test:
	#go test ./
	#go test sbox.go  sbox_test.go twitter.go docker.go 
	go build -o sbox_tester sbox_tester.go twitter.go
	./sbox_tester

clean:
	-docker rmi $(MODULE) localhost:5000/$(MODULE)
	-rm $(MODULE)

install:
	kubectl apply -f deploy.yaml

uninstall:
	kubectl delete -f deploy.yaml

