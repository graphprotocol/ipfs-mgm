FROM golang:1.18 AS builder

RUN mkdir src/app
WORKDIR /src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ipfs-mgm cmd/cli/ipfs-mgm.go

FROM golang:1.18-alpine
COPY --from=builder /bin/ipfs-mgm /bin/ipfs-mgm
ENTRYPOINT [ "/bin/ipfs-mgm" ]
