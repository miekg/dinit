FROM alpine:3.1
MAINTAINER Miek Gieben <miek@miek.nl> (@miekg)

ENV DINIT_TEST 7000
ADD dinit /dinit

ENTRYPOINT [ "/dinit", "-r", "/bin/sleep", "$DINIT_TEST" ]
