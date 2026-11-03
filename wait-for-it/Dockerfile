FROM alpine:3.24@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

ENV WAITFORIT_VERSION="v2.4.1"
RUN wget -q -O /usr/local/bin/waitforit https://github.com/maxcnunes/waitforit/releases/download/${WAITFORIT_VERSION:?}/waitforit-linux_amd64 \
  && chmod +x /usr/local/bin/waitforit \
  && addgroup -S app && adduser -S -g app app

USER app

ENTRYPOINT ["/usr/local/bin/waitforit"]
