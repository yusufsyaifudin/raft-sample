FROM golang:1.14-alpine

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux \
    go build -mod=mod -o raftsample ysf/raftsample/cmd/api


FROM scratch
COPY --from=0 /app/raftsample /raftsample
CMD ["/raftsample"]
