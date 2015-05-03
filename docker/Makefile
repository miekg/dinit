zombie: zombie.c
	gcc -o zombie zombie.c

dinit:
	cp ../dinit .

.PHONY:
docker: zombie dinit
	docker build -t miek/dinit .

.PHONY:
clean:
	rm dinit zombie
