linux_agent_img := slankdev/linux-agent:develop
linux-agent-build:
	CGO_ENABLED=0 go build -o bin/linux-agent cmd/linux-agent/main.go
linux-agent-docker-build: ## Build docker image with the linux-agent.
	docker build -t ${linux_agent_img} -f cmd/linux-agent/Dockerfile .
linux-agent-docker-push: linux-agent-docker-build ## Push docker image with the linux-agent.
	docker push ${linux_agent_img}
linux-agent-run: linux-agent-build
	./bin/linux-agent agent --run-path /usr/local/var/run/linux-agent
