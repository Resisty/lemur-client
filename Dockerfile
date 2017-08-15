FROM scratch

COPY bin/lemur-client /lemur-client
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY config.yaml /
COPY js /js
COPY templates /templates

CMD ["/lemur-client"]
