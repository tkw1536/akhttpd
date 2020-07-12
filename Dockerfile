# build the server
FROM golang as build
ADD . /go/src/akhttpd
WORKDIR /go/src/akhttpd
RUN CGO_ENABLED=0 GOOS=linux make akhttpd

# add it into a scratch image
FROM scratch
WORKDIR /

COPY --from=build /go/src/akhttpd/akhttpd /akhttpd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# by default we use port 8080
# and a limited user, 82 is the convention for www-data
EXPOSE 8080
USER 82:82

# run akhttpd on the chosen port
CMD ["/akhttpd", "0.0.0.0:8080"]
