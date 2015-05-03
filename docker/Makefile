zombie: zombie.c
	gcc -o zombie zombie.c

dinit:
	cp ../dinit .

.PHONY:
image: zombie dinit
	docker build -t $$LOGNAME/dinit .

.PHONY:
clean:
	rm dinit zombie
