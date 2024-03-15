package data

import (
	"context"
	"errors"
	"go-chatter/db"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	Role      string             `bson:"role"`
	CreatedAt time.Time          `bson:"createdAt"`
}

func (u *User) InsertUser(user User) (primitive.ObjectID, error) {
	var collection *mongo.Collection = db.Database.Collection("users")

	// Check if the email already exists in the database
	existingUser, err := GetUserByEmail(user.Email)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if existingUser != nil {
		return primitive.NilObjectID, errors.New("user with this email already exists")
	}

	// Insert the user into the database
	insertedID, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// Extract the inserted ID
	insertedIDPrimitive, ok := insertedID.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, errors.New("failed to assert InsertedID to primitive.ObjectID")
	}

	return insertedIDPrimitive, nil
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	collection := db.Database.Collection("users")
	err := collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

func GetAllUsersWithPagination(pageNumber, pageSize int) ([]User, error) {
	// Declare users slice with pre-defined capacity
	users := make([]User, 0, pageSize)

	// Set options for pagination
	opts := options.Find().
		SetSkip(int64((pageNumber - 1) * pageSize)).
		SetLimit(int64(pageSize))

	// Get collection
	collection := db.Database.Collection("users")

	// Find users with pagination
	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Decode users from cursor
	err = cursor.All(context.Background(), &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserByID(id primitive.ObjectID) (*User, error) {
	var user User
	collection := db.Database.Collection("users")
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

func (u *User) UpdateEmail(newEmail string) error {
	_, err := db.Database.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": u.ID},
		bson.M{"$set": bson.M{"email": newEmail}},
	)
	return err
}

func (u *User) UpdateName(newName string) error {
	_, err := db.Database.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": u.ID},
		bson.M{"$set": bson.M{"name": newName}},
	)
	return err
}

func (u *User) UpdatePassword(newPassword string) error {
	_, err := db.Database.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": u.ID},
		bson.M{"$set": bson.M{"password": newPassword}},
	)
	return err
}

func (u *User) Delete() error {
	collection := db.Database.Collection("users")
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": u.ID})
	if err != nil {
		return err
	}
	return nil
}
