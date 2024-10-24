package backend

import (
	"context"
	"db"
	"fmt"
)

type BikeServer struct {
	UnimplementedBikeServiceServer
	PrismaClient *db.PrismaClient
}

func (server *BikeServer) GetBike(ctx context.Context, req *GetBikeRequest) (*Bike, error) {
	rental, err := server.PrismaClient.Bike.FindUnique(
		db.Bike.ID.Equals(int(req.Id)),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &Bike{
		Id:     int32(rental.ID),
		Model:  rental.Model,
		Status: rental.Status,
	}, nil
}

func (server *BikeServer) CreateBike(ctx context.Context, req *CreateBikeRequest) (*Bike, error) {
	rental, err := server.PrismaClient.Bike.CreateOne(
		db.Bike.Model.Set(req.Model),
		db.Bike.Status.Set(req.Status),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &Bike{
		Id: int32(rental.ID),
	}, nil
}

func (server *BikeServer) UpdateBike(ctx context.Context, req *UpdateBikeRequest) (*Bike, error) {
	rental, err := server.PrismaClient.Bike.FindUnique(
		db.Bike.ID.Equals(int(req.Id)),
	).Update(
		db.Bike.Model.Set(req.Model),
		db.Bike.Status.Set(req.Status),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &Bike{
		Id:     int32(rental.ID),
		Model:  rental.Model,
		Status: req.Status,
	}, nil
}

func (server *BikeServer) DeleteBike(ctx context.Context, req *DeleteBikeRequest) (*DeletedBikeResponse, error) {
	rental, err := server.PrismaClient.Bike.FindUnique(
		db.Bike.ID.Equals(int(req.Id)),
	).Delete().Exec(ctx)
	if err != nil {
		return nil, err
	}
	message := fmt.Sprintf("Bike %d was deleted!", rental.ID)
	return &DeletedBikeResponse{
		Messsage: message,
	}, nil
}

func (server *BikeServer) ListBikes(ctx context.Context, req *ListBikesRequest) (*ListBikesResponse, error) {
	selected, err := server.PrismaClient.Bike.FindMany().Take(int(req.PageSize)).Skip((int(req.Page) - 1) * int(req.PageSize)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	var bikes []*Bike
	for _, bike := range selected {
		bikes = append(bikes, &Bike{
			Id:     int32(bike.ID),
			Model:  bike.Model,
			Status: bike.Status,
		},
		)
	}
	return &ListBikesResponse{
		Bikes: bikes,
	}, nil
}
