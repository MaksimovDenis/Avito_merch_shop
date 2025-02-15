FROM golang:1.22 AS compiling_stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o app ./main.go

FROM alpine:latest

WORKDIR /root/

RUN apk add --no-cache libc6-compat

COPY --from=compiling_stage /app .

RUN chmod +x /root/app

ENTRYPOINT ["./app"]

EXPOSE 8080
