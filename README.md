# Dinit

Donky- or Docker-init is a small init-like daemon useful for use within Docker
containers. It is partly inspired by
[my_init](https://github.com/phusion/baseimage-docker/blob/master/image/bin/my_init).

Dinit will pass any environment variables through to the programs it is starting.
It will pass signals through to the children it is managing. It will reap any zombies
created by its children. It will *not* restart any of its children when they die.
If one of the programs fails to start dinit will exit with an error.

If program daemonize dinit will loose track of them.
