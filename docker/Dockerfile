FROM progrium/busybox
MAINTAINER Miek Gieben <miek@miek.nl> (@miekg)

ADD dinit dinit
ADD zombie zombie

ENTRYPOINT ["/dinit", "/bin/sleep 80", "/zombie"]
