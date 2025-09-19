FROM golang:1.25.1-alpine as builder
ADD . /src
WORKDIR /src
RUN go build -o /out/main .

FROM scratch
COPY --from=builder /out/main /app/main
ENTRYPOINT ["/app/main"]
