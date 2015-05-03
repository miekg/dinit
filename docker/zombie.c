#include <stdlib.h>
#include <sys/types.h>
#include <unistd.h>
#include <stdio.h>

int main()
{
	pid_t pid;
	int i;

	for (i = 0; i < 100; i++) {
		// Create child.
		pid = fork();
		if (pid > 0) {
			printf("zombie: forked %d\n", pid);
			// Parent process.
			sleep(1);
		} else {
			// Child process. Exit immediately.
			exit(0);
		}
	}
	printf("zombie: done");
	return 0;
}
