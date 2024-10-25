// service/bike_service_grpc_test.go
package main_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"backend"
	pb "backend"
	"db"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func dialer(bikeService pb.BikeServiceServer) func(context.Context, string) (net.Conn, error) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterBikeServiceServer(s, bikeService)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return func(ctx context.Context, s string) (net.Conn, error) {
		return lis.Dial()
	}
}

func setupBikeService(t *testing.T) (pb.BikeServiceClient, *db.PrismaClient, context.Context, func()) {
	// Initialize Prisma client
	prismaClient := db.NewClient()
	err := prismaClient.Connect()
	assert.NoError(t, err)

	// Create BikeService server
	bikeService := &backend.BikeServer{
		PrismaClient: prismaClient,
	}

	// Set up in-memory gRPC connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer(bikeService)), grpc.WithInsecure())
	assert.NoError(t, err)

	// Create BikeService client
	client := pb.NewBikeServiceClient(conn)

	// Return cleanup function
	cleanup := func() {
		conn.Close()
		prismaClient.Disconnect()
	}

	return client, prismaClient, ctx, cleanup
}

func TestCreateBikeGRPC(t *testing.T) {
	client, prismaClient, ctx, cleanup := setupBikeService(t)
	defer cleanup()

	req := &pb.CreateBikeRequest{
		Model:  "Test Bike",
		Status: "available",
	}

	// Act
	res, err := client.CreateBike(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, req.Model, res.Model)
	assert.Equal(t, req.Status, res.Status)

	// Clean up
	_, err = prismaClient.Bike.FindUnique(
		db.Bike.ID.Equals(int(res.Id)),
	).Delete().Exec(ctx)
	assert.NoError(t, err)
}

func TestGetBikeGRPC(t *testing.T) {
	client, prismaClient, ctx, cleanup := setupBikeService(t)
	defer cleanup()

	// First, create a bike to retrieve
	createdBike, err := prismaClient.Bike.CreateOne(
		db.Bike.Model.Set("Test Bike"),
		db.Bike.Status.Set("available"),
	).Exec(ctx)
	assert.NoError(t, err)

	req := &pb.GetBikeRequest{
		Id: int32(createdBike.ID),
	}

	// Act
	res, err := client.GetBike(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, createdBike.ID, int(res.Id))
	assert.Equal(t, createdBike.Model, res.Model)
	assert.Equal(t, createdBike.Status, res.Status)

	// Clean up
	_, err = prismaClient.Bike.FindUnique(
		db.Bike.ID.Equals(createdBike.ID),
	).Delete().Exec(ctx)
	assert.NoError(t, err)
}

func TestUpdateBikeGRPC(t *testing.T) {
	client, prismaClient, ctx, cleanup := setupBikeService(t)
	defer cleanup()

	// Create a bike to update
	createdBike, err := prismaClient.Bike.CreateOne(
		db.Bike.Model.Set("Old Model"),
		db.Bike.Status.Set("available"),
	).Exec(ctx)
	assert.NoError(t, err)

	req := &pb.UpdateBikeRequest{
		Id:     int32(createdBike.ID),
		Model:  "New Model",
		Status: "in_service",
	}

	// Act
	res, err := client.UpdateBike(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, req.Id, res.Id)
	assert.Equal(t, req.Model, res.Model)
	assert.Equal(t, req.Status, res.Status)

	// Clean up
	_, err = prismaClient.Bike.FindUnique(
		db.Bike.ID.Equals(createdBike.ID),
	).Delete().Exec(ctx)
	assert.NoError(t, err)
}

func TestDeleteBikeGRPC(t *testing.T) {
	client, prismaClient, ctx, cleanup := setupBikeService(t)
	defer cleanup()

	// Create a bike to delete
	createdBike, err := prismaClient.Bike.CreateOne(
		db.Bike.Model.Set("Test Bike"),
		db.Bike.Status.Set("available"),
	).Exec(ctx)
	assert.NoError(t, err)

	req := &pb.DeleteBikeRequest{
		Id: int32(createdBike.ID),
	}

	// Act
	res, err := client.DeleteBike(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("Bike %d was deleted!", req.Id), res.Messsage)

	// Verify deletion
	_, err = prismaClient.Bike.FindUnique(
		db.Bike.ID.Equals(createdBike.ID),
	).Exec(ctx)
	assert.Error(t, err) // Should return an error because the bike no longer exists
}

func TestListBikesGRPC(t *testing.T) {
	client, prismaClient, ctx, cleanup := setupBikeService(t)
	defer cleanup()

	// Create multiple bikes
	bikeModels := []string{"Bike A", "Bike B", "Bike C"}
	for _, model := range bikeModels {
		_, err := prismaClient.Bike.CreateOne(
			db.Bike.Model.Set(model),
			db.Bike.Status.Set("available"),
		).Exec(ctx)
		assert.NoError(t, err)
	}

	req := &pb.ListBikesRequest{
		Page:     1,
		PageSize: 10,
	}

	// Act
	res, err := client.ListBikes(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res.Bikes), 3)

	// Clean up
	assert.NoError(t, err)
}
