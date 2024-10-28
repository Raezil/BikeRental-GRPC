# BikeRental-GRPC

Sample biking rental service using grpc.

## Run Docker
```
docker build -t app8:latest .
docker run -d -p 50051:50051 -p 8080:8080 --name app8 app8:latest
```
