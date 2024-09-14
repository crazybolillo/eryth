FROM golang:1.22-alpine

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
RUN go build -o /bin/eryth ./cmd

FROM scratch

COPY --from=0 /bin/eryth /bin/eryth

EXPOSE 8080

ENTRYPOINT ["eryth"]
