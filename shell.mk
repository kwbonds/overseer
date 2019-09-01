.PHONY: shell-build-overseer
shell-build-overseer:
	bash scripts/docker-build-hub.sh overseer

.PHONY: shell-build-overseer-webhook-bridge
shell-build-overseer-webhook-bridge:
	bash scripts/docker-build-hub.sh overseer-webhook-bridge Dockerfile.webhook-bridge