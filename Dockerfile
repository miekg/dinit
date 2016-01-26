FROM alpine:3.1
MAINTAINER Miek Gieben <miek@miek.nl> (@miekg)

ENV DINIT_TEST 70
ADD dinit /dinit
ADD zombie.sh /zombie.sh

ENTRYPOINT [ "/dinit", "-r", "/zombie.sh", "-r", "/bin/sleep", "$DINIT_TEST" ]
