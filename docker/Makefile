zombie: zombie.c
	gcc -o zombie zombie.c

dinit:
	cp ../dinit .

.PHONY:
build: zombie dinit
	docker build -t $$LOGNAME/dinit .

.PHONY:
rmi:
	docker rmi -f $$LOGNAME/dinit:latest

.PHONY:
clean:
	rm dinit zombie
