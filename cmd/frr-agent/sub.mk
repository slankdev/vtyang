frr_agent_img := slankdev/frr-agent:develop
frr-agent-build:
	CGO_ENABLED=0 go build -o bin/frr-agent cmd/frr-agent/main.go
frr-agent-docker-build: ## Build docker image with the frr-agent.
	docker build -t ${frr_agent_img} -f cmd/frr-agent/Dockerfile .
frr-agent-docker-push: frr-agent-docker-build ## Push docker image with the frr-agent.
	docker push ${frr_agent_img}
frr-agent-run: frr-agent-build
	./bin/frr-agent agent --run-path /usr/local/var/run/frr-agent
