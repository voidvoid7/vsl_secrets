# Build stage
FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod ./
RUN mkdir /app/bin

# copy source files
COPY cmd ./cmd
COPY internal ./internal

# disable cgo for static linking
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp cmd/main.go
### 
# Final image
FROM scratch

WORKDIR /app/bin
EXPOSE 8080

COPY ./zz__dockerconfig/passwd.nobody /etc/passwd
COPY --from=builder /app/myapp /app/bin/myapp

USER nobody

ENTRYPOINT ["/app/bin/myapp"]