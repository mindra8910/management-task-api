FROM golang:alpine

# Install git (jika belum)
RUN apk add --no-cache git

# Install Air
RUN go install github.com/air-verse/air@latest

ENV PATH="/go/bin:${PATH}"

WORKDIR /app

# Copy source code (tidak perlu go mod download)
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]