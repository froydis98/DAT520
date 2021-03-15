# How to run on docker

## Creating a common network

So that the servers and client can find each other we need to configure a common network. This is only necessary to do the first time.
I have choose that the network will have the ip addresses within the subnet 192.168.0.0/16. 

To configure this on your machine, go to the root (*Fr-ydis-Simon*) and run the command: 
```
docker network create --subnet=192.168.0.0/16 dat520-network
```

## The servers
To create the docker image for the server:
(Everything is done from the root *Fr-ydis-Simon*)

Build the docker image:
```
docker build -t server -f .\lab4\Dockerfile .
```

Run the docker image:
```
docker run -itd --name server1 --network=dat520-network --ip=192.168.1.2 --rm server
```
This will start up the server. Since we need to write in the index of the server we want to attach the server output to powershell (or similar). To do this we write:
```
docker attach server1
```
When this is attached, write **0** as input to start the server.

Do the same to the net two servers:
Server 2:
```
docker run -itd --name server2 --network=dat520-network --ip=192.168.1.3 --rm server
```
Followed by: **docker attach server2** and the input **1**.

Server 3:
```
docker run -itd --name server3 --network=dat520-network --ip=192.168.1.4 --rm server
```
Followed by: **docker attach server3** and the input **2**.

## The clients

We need to build the clients dockerfile as well. Same as with the servers, from root.

```
docker build -t client -f .\lab4\client\Dockerfile .
```

Client 1:
```
docker run -itd --name client1 --network=dat520-network --ip=192.168.2.2 --rm client
```

attach it and give it input **0**.

Client 2:
```
docker run -itd --name client2 --network=dat520-network --ip=192.168.2.3 --rm client
```

attach it and give it input **1**.
