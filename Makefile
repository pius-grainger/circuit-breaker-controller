# Docker Hub configuration
DOCKER_HUB_USER ?= your-dockerhub-username
IMG ?= $(DOCKER_HUB_USER)/circuit-breaker-controller:latest
LOCAL_IMG ?= circuit-breaker-controller:latest

.PHONY: deps
deps:
	go mod tidy

.PHONY: build
build: deps
	go build -o bin/manager main.go

.PHONY: run
run: deps
	go run ./main.go

.PHONY: docker-build
docker-build:
	eval $$(minikube docker-env) && docker build -t ${LOCAL_IMG} .

.PHONY: docker-load
docker-load: docker-build
	minikube image load ${LOCAL_IMG}

.PHONY: docker-build-hub
docker-build-hub:
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: docker-build-hub
	docker push ${IMG}

.PHONY: docker-login
docker-login:
	docker login

.PHONY: install
install:
	kubectl apply -f config/crd/ --validate=false

.PHONY: uninstall
uninstall:
	kubectl delete -f config/crd/

.PHONY: deploy-rbac
deploy-rbac:
	kubectl create namespace circuit-breaker-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f config/rbac/

.PHONY: sample
sample:
	kubectl apply -f config/samples/

.PHONY: helm-install
helm-install:
	kubectl delete crd circuitbreakers.circuitbreaker.io --ignore-not-found
	helm install circuit-breaker-controller helm/circuit-breaker-controller/

.PHONY: helm-uninstall
helm-uninstall:
	helm uninstall circuit-breaker-controller