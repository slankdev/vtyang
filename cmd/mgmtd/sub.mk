mgmtd_img := slankdev/mgmtd:develop
mgmtd-build:
	GOOS=linux CGO_ENABLED=0 go build -o bin/mgmtd cmd/mgmtd/main.go
mgmtd-docker-build: ## Build docker image with the mgmtd.
	docker build -t ${mgmtd_img} -f cmd/mgmtd/Dockerfile .
mgmtd-docker-push: mgmtd-docker-build ## Push docker image with the mgmtd.
	docker push ${mgmtd_img}
mgmtd-run: mgmtd-build
	./bin/mgmtd agent --run-path /usr/local/var/run/mgmtd
