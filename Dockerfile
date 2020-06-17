# build the server
FROM golang as build
ADD *.go /app/
ADD go.mod /app/go.mod
ADD go.sum /app/go.sum
WORKDIR /app/
RUN go get -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app/akhttpd .

# add it into a scratch image
FROM scratch
WORKDIR /

COPY --from=build /app/akhttpd /akhttpd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# and set the entry command
EXPOSE 80

CMD ["/akhttpd", "0.0.0.0:80"]
