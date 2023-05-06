# Mychat

### My chat server made with golang.  
Created for my docker study.  
You can use images from onetwohour/chat-server, onetwohour/chat-client at docker hub.  
Use docker-compose file to create a chat-server.  
```ubuntu
docker-compose up -d
docker run -it --rm --network $(id -un)_server-client-net onetwohour/chat-client:0.2
