FROM golang:1.25.5-alpine AS builder

ARG TARGETARCH

WORKDIR /app
COPY go.mod go.sum .
RUN go mod download && apk add upx

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -trimpath -o epay-bot . && upx -9 epay-bot

FROM alpine

WORKDIR /app/

COPY --from=builder /app/epay-bot .

CMD ["./epay-bot"]
