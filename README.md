# whatsapp-ki-maya

## Usage
### Using Docker
- Build the docker image
```bash
docker build -t whatsapp-ki-maya .
```
- Run the docker container
```bash
docker run -d -p 8050:8050 --name whatsapp-ki-maya whatsapp-ki-maya
```

### Login to whatsapp
- Open the browser and navigate to `http://localhost:8050/login`
- Scan the QR code using your phone
- You are now logged in to whatsapp

### Connecting to Websocket
- Using Postman
  - Create a new request
  - Set the request type to `WebSocket`
  - Set the request URL to `ws://localhost:8050/ws?sub=$chat_id`
  - Click on `Connect` button
  - Send the first message as API token
  - You are now connected to the bot