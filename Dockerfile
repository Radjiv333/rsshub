FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Copy all the files into the container
COPY . .

# Build the Go binary for your app
RUN go build -o rsshub ./cmd/main.go

# Install migrate tool only if necessary (optional if you plan to run it directly from your app)
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Entry point or command to run migration and start the app
CMD ["sh", "-c", "migrate -path=./migrations -database \"postgres://postgres:changeme@db:5432/rsshub?sslmode=disable\" up && ./rsshub fetch"]
