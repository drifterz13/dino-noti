FROM golang:1.22-alpine as builder

ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/dino-noti ./main.go

# Expose the port if running as a service (optional for JOB)
# EXPOSE 8080

# Command to run the application
# For a Cloud Run JOB, this executable will run and then the container stops.
# For a Cloud Run SERVICE, this might start an HTTP server, which would then
# be triggered by Cloud Scheduler via HTTP.
ENTRYPOINT ["/app/dino-noti"]
