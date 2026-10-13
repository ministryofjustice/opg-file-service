FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

ENV WAITFORIT_VERSION="v2.4.1"
RUN wget -q -O /usr/local/bin/waitforit https://github.com/maxcnunes/waitforit/releases/download/${WAITFORIT_VERSION:?}/waitforit-linux_amd64 \
  && chmod +x /usr/local/bin/waitforit \
  && addgroup -S app && adduser -S -g app app

USER app

ENTRYPOINT ["/usr/local/bin/waitforit"]
