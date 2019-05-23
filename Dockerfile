############################
# Taken from https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
# https://github.com/chemidy/smallest-secured-golang-docker-image/blob/master/Dockerfile
# STEP 1 build executable binary
############################
FROM golang:alpine as builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

RUN adduser -D -g '' appuser

WORKDIR $GOPATH/src/PasswordUsage
COPY ./src .

RUN go get -d -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o password-usage-checker

############################
# STEP 2 build a small image
############################
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/PasswordUsage/password-usage-checker password-usage-checker

USER appuser

ENTRYPOINT ["./password-usage-checker"]