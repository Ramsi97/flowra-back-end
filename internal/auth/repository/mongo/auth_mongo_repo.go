package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/auth/domain"
	"github.com/Ramsi97/flowra-back-end/internal/auth/repository/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userDocument struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	FullName          string             `bson:"full_name"`
	Email             string             `bson:"email"`
	Password          string             `bson:"password"`
	Gender            string             `bson:"gender"`
	ProfilePictureURL string             `bson:"profile_picture_url"`
	RestDays          []int              `bson:"rest_days"`
	WorkDayStart      string             `bson:"work_day_start"`
	WorkDayEnd        string             `bson:"work_day_end"`
	CreatedAt         time.Time          `bson:"created_at"`
}

type authMongoRepo struct {
	collection *mongo.Collection
}

func NewAuthMongoRepo(db *mongo.Database) interfaces.AuthRepository {
	return &authMongoRepo{
		collection: db.Collection("users"),
	}
}

func (r *authMongoRepo) CreateUser(ctx context.Context, user *domain.User) error {
	doc := userDocument{
		ID:                primitive.NewObjectID(),
		FullName:          user.FullName,
		Email:             user.Email,
		Password:          user.Password,
		Gender:            user.Gender,
		ProfilePictureURL: user.ProfilePictureURL,
		RestDays:          user.RestDays,
		WorkDayStart:      user.WorkDayStart,
		WorkDayEnd:        user.WorkDayEnd,
		CreatedAt:         user.CreatedAt,
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	user.ID = doc.ID.Hex()
	return nil
}

func (r *authMongoRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var doc userDocument
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.User{
		ID:                doc.ID.Hex(),
		FullName:          doc.FullName,
		Email:             doc.Email,
		Password:          doc.Password,
		Gender:            doc.Gender,
		ProfilePictureURL: doc.ProfilePictureURL,
		RestDays:          doc.RestDays,
		WorkDayStart:      doc.WorkDayStart,
		WorkDayEnd:        doc.WorkDayEnd,
		CreatedAt:         doc.CreatedAt,
	}, nil
}

func (r *authMongoRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var doc userDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.User{
		ID:                doc.ID.Hex(),
		FullName:          doc.FullName,
		Email:             doc.Email,
		Password:          doc.Password,
		Gender:            doc.Gender,
		ProfilePictureURL: doc.ProfilePictureURL,
		RestDays:          doc.RestDays,
		WorkDayStart:      doc.WorkDayStart,
		WorkDayEnd:        doc.WorkDayEnd,
		CreatedAt:         doc.CreatedAt,
	}, nil
}
