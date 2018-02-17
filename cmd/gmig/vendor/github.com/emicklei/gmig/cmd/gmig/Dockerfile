FROM golang:1.9.2

RUN apt-get install -y ca-certificates
RUN update-ca-certificates

WORKDIR /go/src/github.com/kramphub/kiya/
COPY . .
ARG version
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=$version" .


FROM alpine
COPY --from=0 /go/src/github.com/emicklei/gmig /usr/bin/
COPY --from=0 /etc/ssl/certs/ /etc/ssl/certs/

ENTRYPOINT ["kiya"]