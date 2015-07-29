FROM alpine:3.1

ENV HALLO MIEK

ADD dinit /dinit

ENTRYPOINT [ "/dinit", "echo $HALLO" ]
