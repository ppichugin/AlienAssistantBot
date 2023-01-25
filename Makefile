# Variables
DOCKER_COMPOSE = docker-compose
APP_NAME = alien-assistant-bot
TAG_NAME = petrodev/$(APP_NAME)
VERSION = 0.2.0

# Targets
## Build backend of the application and pull it to Docker Hub
build-push:
	docker build -f Dockerfile.backend . -t $(TAG_NAME):$(VERSION)
	docker image tag $(TAG_NAME):$(VERSION) $(TAG_NAME):latest
	docker image push --all-tags $(TAG_NAME)
	docker rmi $(docker images -f “dangling=true” -q)

## Run containers with the image pulled from Docker Hub
run-pull:
	$(DOCKER_COMPOSE) -f docker-compose.yml up -d

## Run containers with full local re-build of backend part
run-rebuild:
	$(DOCKER_COMPOSE) -f docker-compose.full-rebuild.yml up -d --build
	#docker rmi $(docker images -f “dangling=true” -q)

## Stop containers and keep volumes
stop:
	$(DOCKER_COMPOSE) down

## Stop containers and remove volumes
down:
	$(DOCKER_COMPOSE) down --volumes

## Enter to running DB container
exec-db:
	docker exec -it tg-bot-db bash -c "PGPASSWORD=pg psql -h db -U pg -d secretsdbbot"