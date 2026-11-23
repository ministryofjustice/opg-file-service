FROM alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

ENV WAITFORIT_VERSION="v2.4.1"
RUN wget -q -O /usr/local/bin/waitforit https://github.com/maxcnunes/waitforit/releases/download/${WAITFORIT_VERSION:?}/waitforit-linux_amd64 \
  && chmod +x /usr/local/bin/waitforit \
  && addgroup -S app && adduser -S -g app app

USER app

ENTRYPOINT ["/usr/local/bin/waitforit"]
