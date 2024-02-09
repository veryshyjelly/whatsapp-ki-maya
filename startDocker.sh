sudo docker build -t whatsapp-ki-maya .
sudo docker stop whatsapp-ki-maya && sudo docker rm whatsapp-ki-maya
sudo docker run -d -p 8050:8050 --name whatsapp-ki-maya whatsapp-ki-maya