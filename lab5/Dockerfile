FROM golang

# Copy dependencies from your lab3 files to the container's workspace.
COPY ./lab3 /go/src/app/lab3

# Copy the local package files (lab4) to the container's workspace.
COPY . /go/src/app
COPY ./lab5/netConfig.json /go/src/app/netConfig.json
COPY ./lab5/clientNetConfig.json /go/src/app/clientNetConfig.json

WORKDIR /go/src/app

#Build and install your application inside the container.
RUN go install -v ./lab5/server

ENTRYPOINT ["/go/bin/server"]
