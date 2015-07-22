all: dinit

dinit: env.go main.go process.go
	CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo

.PHONY: clean
clean:
	rm -f dinit
