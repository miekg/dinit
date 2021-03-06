all: dinit

dinit: env.go main.go process.go unix.go arg.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -a -tags netgo -installsuffix netgo -o dinit

.PHONY: docker
docker:
	docker build -t $$USER/dinit .

.PHONY: clean
clean:
	rm -f dinit
