DOCKER_VOL := $(CURDIR)/build/data/compost_bin
MYSQL_VOL := $(CURDIR)/build/data/mysql
REDIS_VOL := $(CURDIR)/build/data/redis

default: build

build: build/docker

build/docker: build/binary
	mkdir -p $(DOCKER_VOL)
	mkdir -p $(MYSQL_VOL)
	mkdir -p $(REDIS_VOL)
	docker buildx build -f ./cmd/Dockerfile -t compost-bin:v0.1 .

build/binary:
	mkdir $(CURDIR)/build
	GOOS=linux go build -o ./build ./cmd

run: run/docker

run/docker:
	docker-compose up -d

test: build/docker run/docker
	go clean --testcache
	LOG_DIR="$(CURDIR)/build/data/compost_bin" go test ./test/...

clean:
	docker-compose down
	docker rmi compost-bin:v0.1
	$(RM) -r ./build