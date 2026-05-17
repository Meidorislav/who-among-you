FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/who-among-you

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /server /server
ENTRYPOINT ["/server"]
