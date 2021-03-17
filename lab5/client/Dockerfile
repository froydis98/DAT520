FROM golang

# Copy the local package files (lab4) to the container's workspace.
COPY . /go/src/clientapp
COPY ./lab5/netConfig.json /go/src/clientapp/netConfig.json
COPY ./lab5/clientNetConfig.json /go/src/clientapp/clientNetConfig.json

WORKDIR /go/src/clientapp

#Build and install your application inside the container.
RUN go install -v ./lab5/client

ENTRYPOINT ["/go/bin/client"]