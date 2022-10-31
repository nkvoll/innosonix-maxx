FROM --platform=linux/amd64 golang:1.19 as builder

ENV GOOS linux
ENV GOARCH amd64

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY Makefile Makefile
COPY cmd cmd
COPY pkg pkg
COPY internal internal

RUN make build

# runtime
FROM --platform=linux/amd64 golang:1.19

WORKDIR /
COPY --from=builder /usr/src/app/out/imctl /usr/local/bin/imctl
CMD ["/usr/local/bin/imctl"]