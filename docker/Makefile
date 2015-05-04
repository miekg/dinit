zombie: zombie.c
	gcc -o zombie zombie.c

dinit:
	cp ../dinit .

.PHONY:
build: zombie dinit
	docker build -t $$LOGNAME/dinit . && docker images

.PHONY:
run:
	docker run -e DINIT_VERBOSE=true $$LOGNAME/dinit:latest "/bin/sleep 80" "/zombie"

.PHONY:
rmi:
	docker rmi -f $$LOGNAME/dinit:latest

.PHONY:
clean:
	rm dinit zombie
