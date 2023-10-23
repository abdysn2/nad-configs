FROM golang:1.19-alpine3.15
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/function ./
FROM gcr.io/distroless/static:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
COPY cni-schemas/ /usr/local/bin/cni-schemas/
ENTRYPOINT ["function"]
