FROM golang:1.20.4

# Set destination for COPY command
WORKDIR /app

# Download go modules
COPY go.mod go.sum ./
RUN go mod download
COPY app/*.go ./
COPY ./db/items.db ./db/items.db


# Create image directory
RUN mkdir -p /app/images

RUN CGO_ENABLED=1 GOOS=linux go build -o /mercari

CMD ["/mercari"]
