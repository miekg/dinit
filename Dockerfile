FROM alpine:3.1
MAINTAINER Miek Gieben <miek@miek.nl> (@miekg)

ENV DINIT_TEST miek
ADD dinit /dinit

ENTRYPOINT [ "/dinit", "-r", "/bin/echo", "$DINIT_TEST" ]
