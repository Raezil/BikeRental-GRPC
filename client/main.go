package main

import (
	. "backend"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := NewAuthClient(conn)
	loginReply, err := client.Login(context.Background(), &LoginRequest{
		Email:    "kmosc@example.com",
		Password: "password",
	})
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	token := loginReply.Token
	fmt.Println("Received JWT token:", token)
	md := metadata.Pairs("authorization", token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	protectedReply, err := client.SampleProtected(ctx, &ProtectedRequest{
		Text: "Hello from client",
	})
	if err != nil {
		log.Fatalf("SampleProtected failed: %v", err)
	}
	fmt.Println("SampleProtected response:", protectedReply.Result)
	bikeClient := NewBikeServiceClient(conn)
	bikeReply, err := bikeClient.CreateBike(ctx, &CreateBikeRequest{
		Model:  "Bike X",
		Status: "ONGOING",
	})
	if err != nil {
		log.Fatalf("CreateBike failed: %v", err)
	}
	bikeReply, err = bikeClient.GetBike(ctx, &GetBikeRequest{
		Id: bikeReply.Id,
	})
	if err != nil {
		panic(err)
	}
	bikeReply, err = bikeClient.UpdateBike(ctx, &UpdateBikeRequest{
		Id:     bikeReply.Id,
		Status: "COMPLETED",
	})
	rentalClient := NewRentalServiceClient(conn)

	rentalReply, err := rentalClient.CreateRental(ctx, &CreateRentalRequest{
		BikeId: bikeReply.Id,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rentalReply)

	rentalUpdateReply, err := rentalClient.UpdateRental(ctx, &UpdateRentalRequest{
		Id:      rentalReply.Id,
		BikeId:  rentalReply.BikeId,
		EndTime: timestamppb.New(time.Now()),
		Status:  "DONE",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rentalUpdateReply)

}
