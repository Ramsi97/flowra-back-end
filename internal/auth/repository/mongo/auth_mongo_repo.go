package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/auth/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// userDocument is the MongoDB representation of a user.
// We store the password hash here, not in the domain User.
type userDocument struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	FullName       string             `bson:"full_name"`
	Email          string             `bson:"email"`
	HashedPassword string             `bson:"hashed_password"`
	Gender         string             `bson:"gender"`
	ProfilePicture string             `bson:"profile_picture,omitempty"`
	CreatedAt      time.Time          `bson:"created_at"`
}

type authMongoRepo struct {
	collection *mongo.Collection
}

// NewAuthMongoRepo creates a new MongoDB-backed AuthRepository.
func NewAuthMongoRepo(db *mongo.Database) *authMongoRepo {
	return &authMongoRepo{
		collection: db.Collection("users"),
	}
}

// CreateUser inserts a new user document into MongoDB.
// The user.Password field is expected to already be hashed.
func (r *authMongoRepo) CreateUser(ctx context.Context, user *domain.User) error {
	doc := userDocument{
		ID:             primitive.NewObjectID(),
		FullName:       user.FullName,
		Email:          user.Email,
		HashedPassword: user.Password, // already hashed by usecase
		Gender:         user.Gender,
		CreatedAt:      time.Now(),
	}
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

// FindByEmail looks up a user by email address.
// Returns the domain.User with the hashed password in the Password field.
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
		ID:        doc.ID.Hex(),
		FullName:  doc.FullName,
		Email:     doc.Email,
		Password:  doc.HashedPassword,
		Gender:    doc.Gender,
		CreatedAt: doc.CreatedAt,
	}, nil
}
