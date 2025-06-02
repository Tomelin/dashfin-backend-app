FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy

FROM golang:1.24 AS builder2

WORKDIR /app

COPY --from=builder /app .
COPY --from=builder /app/* .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o /manager cmd/main.go


FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder2 /manager .

USER 1000:1000

ENTRYPOINT ["/manager"]

EXPOSE 8080

CMD ["/manager"]

