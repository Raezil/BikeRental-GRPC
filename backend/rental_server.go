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

/*
	curl -X POST http://localhost:8080/v1/rentals \
	  -H 'Content-Type: application/json' \
	  -H 'Authorization: $TOKEN' \
	  -d '{
	        "bike_id": 1
	      }'
*/
func (server *RentalServer) CreateRental(ctx context.Context, req *CreateRentalRequest) (*Rental, error) {
	email, err := CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	user, err := server.PrismaClient.User.FindUnique(
		db.User.Email.Equals(email),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	result, err := server.PrismaClient.Rental.CreateOne(
		db.Rental.User.Link(db.User.ID.Equals(int(user.ID))),
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
		Status: "Ongoing",
		// Add other fields as needed
	}

	return rental, nil
}

/*
	curl -X GET http://localhost:8080/v1/rentals/1 \
	  -H 'Authorization: $TOKEN'
*/
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

/*
	curl -X DELETE http://localhost:8080/v1/rentals/1 \
	  -H 'Authorization: $TOKEN'
*/
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

/*
	curl -X PUT http://localhost:8080/v1/rentals/1 \
	  -H 'Content-Type: application/json' \
	  -H 'Authorization: $TOKEN' \
	  -d '{
	        "bike_id": 2,
	        "end_time": "2023-10-01T15:30:00Z",
	        "status": "completed"
	      }'
*/
func (server *RentalServer) UpdateRental(ctx context.Context, req *UpdateRentalRequest) (*Rental, error) {
	email, err := CurrentUser(ctx)
	user, err := server.PrismaClient.User.FindUnique(
		db.User.Email.Equals(email),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	result, err := server.PrismaClient.Rental.FindUnique(
		db.Rental.ID.Equals(user.ID),
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
		Status:  result.Status,
		// Add other fields as needed
	}, nil
}

/*
	curl -X GET 'http://localhost:8080/v1/rentals?page=1&page_size=10' \
	  -H 'Authorization: $TOKEN'
*/
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
			Status:  rental.Status,
			EndTime: timestamppb.New(end),
		})
	}
	return &ListRentalsResponse{
		Rentals: rentals,
	}, nil
}
