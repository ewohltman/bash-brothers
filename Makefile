.PHONY: fmt lint test build image push deploy

version=v0.0.1

fmt:
	gofmt -s -w . && goimports -local github.com/ewohltman/bash-brothers -w . && go mod tidy

lint: fmt
	golangci-lint run ./...

test:
	go test -v -race -coverprofile=coverage.out ./...
	@ echo "all tests passed"

build:
	CGO_ENABLED=0 go build -o build/package/bash-brothers/bash-brothers bash-brothers.go

image:
	docker image build -t ewohltman/bash-brothers:${version} build/package/bash-brothers

push:
	docker push ewohltman/bash-brothers:${version}
	docker tag ewohltman/bash-brothers:${version} ewohltman/bash-brothers:latest

deploy:
	kubectl apply -f deployments/kubernetes/namespace.yml
	kubectl apply -f deployments/kubernetes/deployment.yml
	kubectl apply -f deployments/kubernetes/service.yml
	kubectl apply -f deployments/kubernetes/ingress.yml
