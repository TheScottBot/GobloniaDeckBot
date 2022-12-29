FROM golang:1.16-alpine

# # Set destination for COPY
# WORKDIR /Users/uu/Documents/GitHub/GobloniaDeckBot

# # Download Go modules
# COPY go.mod .
# COPY go.sum .
# RUN go mod download

# # Copy the source code. Note the slash at the end, as explained in
# # https://docs.docker.com/engine/reference/builder/#copy
# COPY *.go ./

# # Build
# RUN go build main.go

# # This is for documentation purposes only.
# # To actually open the port, runtime parameters
# # must be supplied to the docker command.
# EXPOSE 8080

# # (Optional) environment variable that our dockerised
# # application can make use of. The value of environment
# # variables can also be set via parameters supplied
# # to the docker command on the command line.
# #ENV HTTP_PORT=8081

# # Run
# CMD [ "/main" ]

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
WORKDIR $GOPATH/src/github.com/TheScottBot/GobloniaDeckBot

COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["main"]