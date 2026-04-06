package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/task/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type exceptionDocument struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	TaskID       string             `bson:"task_id"`
	Date         string             `bson:"date"`
	NewDuration  string             `bson:"new_duration"`
	NewStartTime *time.Time         `bson:"new_start_time"`
	IsSkipped    bool               `bson:"is_skipped"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

type exceptionMongoRepo struct {
	collection *mongo.Collection
}

func NewExceptionMongoRepo(db *mongo.Database) domain.ExceptionRepository {
	return &exceptionMongoRepo{collection: db.Collection("task_exceptions")}
}

func (r *exceptionMongoRepo) Create(ctx context.Context, ex *domain.TaskException) error {
	doc := exceptionDocument{
		ID:           primitive.NewObjectID(),
		TaskID:       ex.TaskID,
		Date:         ex.Date,
		NewDuration:  ex.NewDuration,
		NewStartTime: ex.NewStartTime,
		IsSkipped:    ex.IsSkipped,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_, err := r.collection.InsertOne(ctx, doc)
	if err == nil {
		ex.ID = doc.ID.Hex()
	}
	return err
}

func (r *exceptionMongoRepo) FindByTaskAndDate(ctx context.Context, taskID, date string) (*domain.TaskException, error) {
	var doc exceptionDocument
	if err := r.collection.FindOne(ctx, bson.M{"task_id": taskID, "date": date}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.TaskException{
		ID:           doc.ID.Hex(),
		TaskID:       doc.TaskID,
		Date:         doc.Date,
		NewDuration:  doc.NewDuration,
		NewStartTime: doc.NewStartTime,
		IsSkipped:    doc.IsSkipped,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}, nil
}

func (r *exceptionMongoRepo) ListByTask(ctx context.Context, taskID string) ([]domain.TaskException, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"task_id": taskID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var exceptions []domain.TaskException
	for cursor.Next(ctx) {
		var doc exceptionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		exceptions = append(exceptions, domain.TaskException{
			ID:           doc.ID.Hex(),
			TaskID:       doc.TaskID,
			Date:         doc.Date,
			NewDuration:  doc.NewDuration,
			NewStartTime: doc.NewStartTime,
			IsSkipped:    doc.IsSkipped,
			CreatedAt:    doc.CreatedAt,
			UpdatedAt:    doc.UpdatedAt,
		})
	}
	return exceptions, cursor.Err()
}

func (r *exceptionMongoRepo) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid exception id")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
