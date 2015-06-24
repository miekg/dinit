all: dinit

dinit: env.go main.go process.go:
	go build -a -tags netgo -installsuffix netgo
