ARG BUILDER_IMAGE=golang:alpine

###
# BUILD
###
FROM ${BUILDER_IMAGE} as builder

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create a user
ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735
# for more details
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /src/

COPY go.mod .

ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

# this copies the code to the src folder defined as WORKDIR above
COPY . .

# Build the app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' -a \
    -o runner .

###
# Copy executable to alpine image
# Note: Here we need a base image that contains the
# tools specified in the sample.yaml. We use alpine
# because it contains a lot of tools out of the box
###

FROM alpine

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static executable
COPY --from=builder /src/runner .
COPY --from=builder /src/sample.yaml .

# Use an unprivileged user
USER appuser:appuser

ENTRYPOINT ["./runner"]
