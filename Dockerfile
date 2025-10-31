FROM --platform=$BUILDPLATFORM golang:1.23.4-alpine3.20@sha256:8d8e54bb0fb1923112e1c8a185da30258aab6fa41063f63683bcab9d360f1a2e AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

RUN apk --no-cache add ca-certificates

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
	go build -trimpath -ldflags "-s -w" -o /out/portscan ./cmd/main.go

FROM scratch AS runtime

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/portscan /portscan

USER 65532:65532
ENTRYPOINT ["/portscan"]
