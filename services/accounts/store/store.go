package store

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

const UsersCollection = "users"

type UsersStore interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Get(ctx context.Context, id primitive.ObjectID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
}

type store struct {
	conn *mongo.Collection
}

func NewUsersStore(dbConn *mongo.Database) UsersStore {
	return &store{conn: dbConn.Collection(UsersCollection)}
}

func (s *store) Create(ctx context.Context, user *User) error {
	result, err := s.conn.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	log.Printf("user created: id=%v\n", result.InsertedID)
	return nil
}

func (s *store) Update(ctx context.Context, user *User) error {
	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"password":   user.Password,
			"updated_at": user.UpdatedAt,
		},
	}

	filter := bson.M{"_id": bson.M{"$eq": user.Id}}

	result, err := s.conn.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	log.Printf("user updated: total=%v\n", result.ModifiedCount)
	return nil
}

func (s *store) Get(ctx context.Context, id primitive.ObjectID) (*User, error) {
	var user User

	err := s.conn.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}

	log.Printf("found user: id=%v\n", id)

	return &user, nil
}

func (s *store) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	err := s.conn.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	log.Printf("found user: email=%v\n", user.Email)

	return &user, nil
}

func (s *store) GetAll(ctx context.Context) ([]*User, error) {

	cursor, err := s.conn.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*User
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}

	log.Printf("list users: total=%v\n", len(users))

	return users, nil
}
