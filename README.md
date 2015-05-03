# Dinit

Docker-init is a small init-like "daemon" (it is not a daemon) useful for use
within Docker containers. It is partly inspired by
[my_init](https://github.com/phusion/baseimage-docker/blob/master/image/bin/my_init).

Dinit will pass any environment variables through to the programs it is starting.
It will pass signals through to the children it is managing. It will reap any zombies
created by its children. It will *not* restart any of its children if they die.

If one of the programs fails to start dinit will exit with an error. If programs
daemonize dinit will loose track of them.

[Prometheus](http://prometheus.io/) is supported to scrape the number of zombies
(zombies_reaped), but only if -port > 0.

# Why?

See <https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/>.
But a simpler solution. Get a standard container image and instead of:

    ENTRYPOINT ["/bin/sleep", "80"]

Do:

    ADD dinit dinit
    ENTRYPOINT ["/dinit", "/bin/sleep 80"]

# TODO

Go Tests. The docker directory contains some files to run a docker container to
test a few things.

# Misc

Build with `go build -ldflags -s` to reduce the size a bit.
