FROM golang:1.23-alpine AS build
WORKDIR /app
COPY backend/go.mod backend/go.sum* ./
RUN go mod download || true
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /out/hehuan-server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/hehuan-server /usr/local/bin/hehuan-server
ENV HEHUAN_HTTP_ADDR=:8080
EXPOSE 8080
CMD ["hehuan-server"]
