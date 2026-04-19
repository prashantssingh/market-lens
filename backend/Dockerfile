FROM golang:1.24-bookworm AS build

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o market-lens-api .

FROM debian:bookworm-slim
WORKDIR /app
ENV DATA_DIR=/data
COPY --from=build /app/market-lens-api /app/market-lens-api
EXPOSE 8080
CMD ["/app/market-lens-api"]
