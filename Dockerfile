FROM golang:1.17-alpine AS build

WORKDIR /app

COPY go.??? ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -v ./...

RUN CGO_ENABLED=0 go build -o /build/rate-limit

FROM alpine:3.14

RUN apk add --no-cache ca-certificates

COPY --from=build /build/rate-limit /app/rate-limit

ENV NETWORK_PREFIX=24
ENV NUMBER_OF_REQUESTS=100
ENV UNIT_TIME=Minute
ENV LIMIT_TIME=1
ENV WAIT_TIME=2
ENV HOST=localhost
ENV PORT=8000

EXPOSE ${PORT}

CMD ["/app/rate-limit"]
