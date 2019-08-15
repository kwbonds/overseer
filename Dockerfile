FROM golang:1.12.4-alpine3.9 as builder

# Install git
# Git is required for fetching the dependencies.
RUN apk update && apk upgrade && \
    apk add --no-cache gcc g++ git ca-certificates && update-ca-certificates

WORKDIR /build
ADD . .

# Build the binary
RUN go build -ldflags "-X main.version=$(git describe --tags 2>/dev/null || echo 'master')" -a -o /go/bin/main .

############################
# STEP 2 build a small image
############################
FROM alpine:3.9

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy our static executable
COPY --from=builder /go/bin/main /go/bin/main
RUN chmod a+x /go/bin/main

ENTRYPOINT ["/go/bin/main"]