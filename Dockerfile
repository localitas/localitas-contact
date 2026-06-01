FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN go build -o /contact-server ./cmd/contact-server

FROM alpine:3.20
COPY --from=builder /contact-server /usr/local/bin/contact-server
ENV CORE_URL=http://localhost:9090
ENTRYPOINT ["contact-server", "--listen", ":9201", "--core-url"]
CMD ["http://localhost:9090"]
