FROM alpine:3.1
MAINTAINER Miek Gieben <miek@miek.nl> (@miekg)

RUN apk --update add netcat-openbsd

ENV DINIT_TEST 7000
ADD dinit /dinit

ENTRYPOINT [ "/dinit", "-r", "/bin/sleep", "$DINIT_TEST" ]
