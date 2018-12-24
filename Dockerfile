FROM scratch

COPY bin/neosproxy-linux-amd64 /usr/sbin/neosproxy

# install ca root certificates for outgoing https calls
# https://curl.haxx.se/docs/caextract.html
# http://blog.codeship.com/building-minimal-docker-containers-for-go-applications/
#ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY files/cacert.pem /etc/ssl/certs/ca-certificates.crt

COPY files/tmp /tmp

VOLUME /htdocs

EXPOSE 80

ENTRYPOINT ["/usr/sbin/neosproxy"]