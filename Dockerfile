FROM golang:1.15-alpine3.12 as build

COPY . /opt/

WORKDIR /opt 

RUN apk add git
RUN ls -la
RUN cd cmd && \
    go build

FROM alpine:3.12 as run
COPY --from=build /opt/cmd/cmd /opt/deedetext-exporter
ENTRYPOINT ["/opt/deedetext-exporter"]
