FROM golang:1.26 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/auth .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/auth /auth

EXPOSE 3000

USER nonroot:nonroot
ENTRYPOINT ["/auth"]
