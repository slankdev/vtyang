vtyang_img := slankdev/vtyang:develop
vtyang-build:
	CGO_ENABLED=0 go build -o bin/vtyang cmd/vtyang/main.go
vtyang-docker-build: ## Build docker image with the vtyang.
	docker build -t ${vtyang_img} -f cmd/vtyang/Dockerfile .
vtyang-docker-push: vtyang-docker-build ## Push docker image with the vtyang.
	docker push ${vtyang_img}
vtyang-run: vtyang-build
	./bin/vtyang agent --run-path /usr/local/var/run/vtyang
