package backend

import (
	"context"
	"db"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type RentalServer struct {
	UnimplementedRentalServiceServer
	PrismaClient *db.PrismaClient
}

func (server *RentalServer) CreateRental(ctx context.Context, req *CreateRentalRequest) (*Rental, error) {
	result, err := server.PrismaClient.Rental.CreateOne(
		db.Rental.User.Link(db.User.ID.Equals(int(req.UserId))),
		db.Rental.Bike.Link(db.Bike.ID.Equals(int(req.BikeId))),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Map the result to your Rental type if necessary
	rental := &Rental{
		Id:     int32(result.ID),
		UserId: int32(result.UserID),
		BikeId: int32(result.BikeID),
		// Add other fields as needed
	}

	return rental, nil
}

func (server *RentalServer) GetRental(ctx context.Context, req *GetRentalRequest) (*Rental, error) {
	result, err := server.PrismaClient.Rental.FindUnique(
		db.Rental.ID.Equals(int(req.Id)),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	t, _ := result.EndTime()
	return &Rental{
		Id:      int32(result.ID),
		UserId:  int32(result.UserID),
		BikeId:  int32(result.BikeID),
		EndTime: timestamppb.New(t),
	}, nil
}

func (server *RentalServer) DeleteRental(ctx context.Context, req *DeleteRentalRequest) (*DeletedRentalResponse, error) {
	rental, err := server.PrismaClient.Rental.FindUnique(
		db.Rental.ID.Equals(int(req.Id)),
	).Delete().Exec(ctx)
	if err != nil {
		return nil, err
	}
	message := fmt.Sprintf("Bike %d was deleted!", rental.ID)
	return &DeletedRentalResponse{
		Message: message,
	}, nil
}

func (server *RentalServer) UpdateRental(ctx context.Context, req *UpdateRentalRequest) (*Rental, error) {
	result, err := server.PrismaClient.Rental.FindUnique(
		db.Rental.ID.Equals(int(req.Id)),
	).Update(
		db.Rental.Status.Set(req.Status),
		db.Rental.EndTime.Set(req.EndTime.AsTime()),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	time, _ := result.EndTime()

	return &Rental{
		Id:      int32(result.ID),
		UserId:  int32(result.UserID),
		BikeId:  int32(result.BikeID),
		EndTime: timestamppb.New(time),
		// Add other fields as needed
	}, nil
}

func (server *RentalServer) ListRentals(ctx context.Context, req *ListRentalsRequest) (*ListRentalsResponse, error) {
	selected, err := server.PrismaClient.Rental.FindMany().Take(int(req.PageSize)).Skip((int(req.Page) - 1) * int(req.PageSize)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	var rentals []*Rental
	for _, rental := range selected {
		end, _ := rental.EndTime()
		rentals = append(rentals, &Rental{
			Id:      int32(rental.ID),
			UserId:  int32(rental.UserID),
			BikeId:  int32(rental.BikeID),
			EndTime: timestamppb.New(end),
		})
	}
	return &ListRentalsResponse{
		Rentals: rentals,
	}, nil
}
