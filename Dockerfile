# syntax=docker/dockerfile:1
FROM golang:1.21 AS build-stage

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build jacobo.tarrio.org/jtweb/cmd/jtserver

FROM scratch
WORKDIR /app
COPY --from=build-stage /build/jtserver ./
VOLUME /data
VOLUME /mysql
EXPOSE 8080
ENTRYPOINT [ "./jtserver", "--server_address=0.0.0.0:8080" ]
