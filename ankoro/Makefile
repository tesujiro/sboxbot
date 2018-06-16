MODULE=ankoro
GOOS=linux

create_secret:
	/bin/echo -n $$HASHTAG >./.HASHTAG
	/bin/echo -n $$TWITTER_CONSUMER_KEY >./.TWITTER_CONSUMER_KEY
	/bin/echo -n $$TWITTER_CONSUMER_SECRET >./.TWITTER_CONSUMER_SECRET
	/bin/echo -n $$TWITTER_ACCESS_TOKEN >./.TWITTER_ACCESS_TOKEN
	/bin/echo -n $$TWITTER_ACCESS_TOKEN_SECRET >./.TWITTER_ACCESS_TOKEN_SECRET
	kubectl create secret generic $(MODULE)-twitter-apikey --from-file=HASHTAG=./.HASHTAG --from-file=TWITTER_CONSUMER_KEY=./.TWITTER_CONSUMER_KEY --from-file=TWITTER_CONSUMER_SECRET=./.TWITTER_CONSUMER_SECRET --from-file=TWITTER_ACCESS_TOKEN=./.TWITTER_ACCESS_TOKEN --from-file=TWITTER_ACCESS_TOKEN_SECRET=./.TWITTER_ACCESS_TOKEN_SECRET
	for key in HASHTAG TWITTER_CONSUMER_KEY TWITTER_CONSUMER_SECRET TWITTER_ACCESS_TOKEN TWITTER_ACCESS_TOKEN_SECRET;do\
		cat ./.$$key ; \
		rm ./.$$key ; \
	done
	
delete_secret:
	kubectl delete secret $(MODULE)-twitter-apikey
	
clean:
	-docker rmi $(MODULE) localhost:5000/$(MODULE)
	-rm $(MODULE)

install:
	kubectl apply -f deploy.yaml

uninstall:
	kubectl delete -f deploy.yaml

logs:
	kubectl logs -f `kubectl get po | awk '/$(MODULE)/{print $$1}'`

gcloud_install:
	kubectl apply -f gcloud_deploy.yaml

gcloud_uninstall:
	kubectl delete -f gcloud_deploy.yaml

switch_local:
	kubectl config use-context docker-for-desktop --insecure-skip-tls-verify=true

switch_gcloud:
	kubectl config use-context gke_anko-robot_asia-east1-a_ankoro-cluster 

