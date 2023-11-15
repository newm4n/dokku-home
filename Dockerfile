FROM golang:1.21
ENV GOPATH /go
ENV GO111MODULE on
ENV GOOS linux
ENV GOARCH amd64

# Prepare all the dirs.
RUN mkdir -p $GOPATH/src/github.com/hamisid/GoStatic
# Copy the build content.
COPY . $GOPATH/src/github.com/hamisid/GoStatic
# Checkout the go-resource to auto generate statics into go codes.
WORKDIR $GOPATH/src/github.com/hamisid/GoStatic
# Compile the proje ct
RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o GoStatic.app cmd/Main.go

# Now use the deployment image.
FROM alpine:latest
ENV GOPATH /go
RUN apk --no-cache add ca-certificates
# Copy the built binary to the new image.
WORKDIR /root/
COPY --from=0 $GOPATH/src/github.com/hamisid/GoStatic/GoStatic.app .
# Expose port.
EXPOSE 8080
# Execute
CMD ["./GoStatic.app"]
