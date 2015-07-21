# Dinit

## Synopsis

    ./dinit [OPTIONS] CMD [CMD...]

## Description

Docker-init is a small init-like "daemon" (it is not a daemon) for use within
Docker containers.

Dinit will pass any environment variables through to the programs it is
starting. It will pass signals (SIGHUP, SIGTERM and SIGINT) through to the
children it is managing. It will *not* restart any of its children if they die.

If one of the programs fails to start dinit will exit with an error. If programs
daemonize dinit will lose track of them.

Docker-init has the concept of a *primary* process which is the *first* process
listed. If that process dies Docker-init will kill the remaining processes and
exit. This allows for cleanups and container restarts.

### Why?

See <https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/>.
But a simpler solution. Get a standard container image and instead of:

    ENTRYPOINT ["/bin/sleep", "80"]

Do:

    ADD dinit dinit
    ENTRYPOINT ["/dinit", "/bin/sleep 80"]

or

    ENTRYPOINT ["/dinit", "/bin/sleep $TIMEOUT"]

Where `$TIMEOUT` will be expanded by `dinit` itself.

The last command in the list given to `dinit` will *also* get the arguments given
to `docker run`, so the above sleep can be rewritten like:

    ENTRYPOINT ["/dinit", "/bin/sleep"]

And then call `docker run .... 80`

## Options

* `maxproc`: set GOMAXPROCS to the number of CPUs on the host multiplied my `maxproc`, typical
  values are 0.5 or 1.0. When 0.0 `dinit` will not set GOMAXPROCS by itself. If GOMAXPROCS is
  *already* set in the environment this does nothing.
* `start`: run a command when starting up. On any failure, `dinit` exits.
* `stop`: run command on exit.
* `timeout`: time in seconds before SIGKILL is send after the SIGTERM has been sent.

## Examples

...

## See Also

Dinit is partly inspired by
[my_init](https://github.com/phusion/baseimage-docker/blob/master/image/bin/my_init).
