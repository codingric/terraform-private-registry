FROM golang:1.21.6-alpine3.18 AS build
WORKDIR /app
COPY go.* .
RUN go mod download
COPY *.go .
RUN go build -o /terraform-private-registry

FROM alpine:3.18
COPY --from=build /terraform-private-registry /terraform-private-registry
EXPOSE 8080
ENTRYPOINT ["/terraform-private-registry"]