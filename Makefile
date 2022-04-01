IMG ?= quay.io/rcampos/vpc-finder:latest

all: bin/imdsv1Mocker bin/vpcFinder

bin/imdsv1Mocker: cmd/imdsv1Mocker.go
	go build -o bin/imdsv1Mocker cmd/imdsv1Mocker.go

bin/vpcFinder: cmd/vpcFinder.go
	go build -o bin/vpcFinder cmd/vpcFinder.go

image:
	docker build -t $(IMG) .

push: image
	docker push $(IMG)

minikube-deploy: push
	kubectl apply -f manifests/rbac.yaml; \
		sleep 1; \
		kubectl apply -f manifests/minikube-vpc-finder.yaml

undeploy:
	kubectl delete -f manifests/

clean:
	rm -f bin/imdsv1Mocker bin/vpcFinder
