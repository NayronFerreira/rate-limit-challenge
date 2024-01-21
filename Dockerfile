FROM golang:1.21 As builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# Path: Dockerfile

FROM scratch

COPY --from=builder /app/main /main
COPY --from=builder /app/.env /.env

EXPOSE 8080

CMD ["/main"]
