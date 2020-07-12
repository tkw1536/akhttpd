# build the server
FROM golang as build
ADD . /go/src/akhttpd
WORKDIR /go/src/akhttpd
RUN CGO_ENABLED=0 GOOS=linux make akhttpd

# add it into a scratch image
FROM scratch
WORKDIR /

COPY --from=build /go/src/akhttpd /akhttpd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# and set the entry command
EXPOSE 80

CMD ["/akhttpd", "0.0.0.0:80"]
