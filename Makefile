MODULE=sbox
GOOS=linux

all: build container deploy_container

deploy_container:
	for p in `kubectl get pod | grep $(MODULE) | awk '{print $$1}'`;do\
		kubectl delete pod $$p;\
	done
	docker rmi $$(docker images -a --filter "dangling=true" -q) -f

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

GCP_PROJECT=anko-robot
GCP_TAG=asia.gcr.io/$(GCP_PROJECT)/$(MODULE)
register_gcloud:
	docker build -t $(GCP_TAG) .
	gcloud docker -- push $(GCP_TAG)

create_secret:
	/bin/echo -n $$HASHTAG >./.HASHTAG
	/bin/echo -n $$TWITTER_CONSUMER_KEY >./.TWITTER_CONSUMER_KEY
	/bin/echo -n $$TWITTER_CONSUMER_SECRET >./.TWITTER_CONSUMER_SECRET
	/bin/echo -n $$TWITTER_ACCESS_TOKEN >./.TWITTER_ACCESS_TOKEN
	/bin/echo -n $$TWITTER_ACCESS_TOKEN_SECRET >./.TWITTER_ACCESS_TOKEN_SECRET
	kubectl create secret generic twitter-apikey --from-file=HASHTAG=./.HASHTAG --from-file=TWITTER_CONSUMER_KEY=./.TWITTER_CONSUMER_KEY --from-file=TWITTER_CONSUMER_SECRET=./.TWITTER_CONSUMER_SECRET --from-file=TWITTER_ACCESS_TOKEN=./.TWITTER_ACCESS_TOKEN --from-file=TWITTER_ACCESS_TOKEN_SECRET=./.TWITTER_ACCESS_TOKEN_SECRET
	for key in HASHTAG TWITTER_CONSUMER_KEY TWITTER_CONSUMER_SECRET TWITTER_ACCESS_TOKEN TWITTER_ACCESS_TOKEN_SECRET;do\
		cat ./.$$key ; \
		rm ./.$$key ; \
	done
	
delete_secret:
	kubectl delete secret twitter-apikey
	
test:
	go test -v ./

test_ankoro:
	go test -v ankoro_test.go twitter.go

test_docker:
	go test -v docker_test.go docker.go
clean:
	-docker rmi $(MODULE) localhost:5000/$(MODULE)
	-rm $(MODULE)

install:
	kubectl apply -f deploy.yaml

uninstall:
	kubectl delete -f deploy.yaml

logs:
	kubectl logs -f `kubectl get po | awk '/$(MODULE)/{print $$1}'`

