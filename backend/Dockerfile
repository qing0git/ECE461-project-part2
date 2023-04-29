# we will use the latest version of node available from the Docker Hub.
FROM golang:1.18

# Create app directory in container image, where your app will live its lifetime.
ENV APP_HOME /app
WORKDIR $APP_HOME

# Github Token
ENV GITHUB_TOKEN=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

# Install app dependencies
COPY go.mod go.sum ./

# Installing the packages while the image is building
RUN go mod download

# Bundle app source, i.e. copying all your required files for the app
# Note: files & folders inside .dockerignore will not be copied.
COPY . .

# Install Python3
RUN apt-get update && apt-get install -y python3 python3-pip git

# Install the GitPython library
RUN pip3 install GitPython

# Build the Go application
RUN go build -o main .

# The app binds to port 8080, so exposing port 8080 to be used by the docker network
EXPOSE 8080

# Set up the entrypoint for the container
ENTRYPOINT ["./main"]
