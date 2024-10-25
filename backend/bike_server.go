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

/*
	curl -X GET http://localhost:8080/v1/bikes/1 \
	  -H 'Authorization: $TOKEN'
*/
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

/*
	curl -X POST http://localhost:8080/v1/bikes \
	  -H 'Content-Type: application/json' \
	  -H 'Authorization: $TOKEN' \
	  -d '{
	        "model": "Mountain Bike",
	        "status": "available"
	      }'
*/
func (server *BikeServer) CreateBike(ctx context.Context, req *CreateBikeRequest) (*Bike, error) {
	rental, err := server.PrismaClient.Bike.CreateOne(
		db.Bike.Model.Set(req.Model),
		db.Bike.Status.Set(req.Status),
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

/*
	curl -X PUT http://localhost:8080/v1/bikes/1 \
	  -H 'Content-Type: application/json' \
	  -H 'Authorization: $TOKEN' \
	  -d '{
	        "model": "Road Bike",
	        "status": "in_service"
	      }'
*/
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

/*
	curl -X DELETE http://localhost:8080/v1/bikes/1 \
	  -H 'Authorization: $TOKEN'
*/
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

/*
	curl -X GET 'http://localhost:8080/v1/bikes?page=1&page_size=10' \
	  -H 'Authorization: $TOKEN'
*/
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
