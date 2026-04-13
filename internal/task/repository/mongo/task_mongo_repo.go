package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/task/domain"
	"github.com/Ramsi97/flowra-back-end/internal/task/repository/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type taskDocument struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      string             `bson:"user_id"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Duration    string             `bson:"duration"`
	Priority    int                `bson:"priority"`
	IsHard      bool               `bson:"is_hard"`
	Status      string             `bson:"status"`
	Deadline    *time.Time         `bson:"deadline"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type taskMongoRepo struct {
	collection *mongo.Collection
}

func NewTaskMongoRepo(db *mongo.Database) interfaces.TaskRepository {
	return &taskMongoRepo{collection: db.Collection("tasks")}
}

func toDoc(t *domain.Task) taskDocument {
	return taskDocument{
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Duration:    t.Duration,
		Priority:    t.Priority,
		IsHard:      t.IsHard,
		Status:      t.Status,
		Deadline:    t.Deadline,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func fromDoc(d taskDocument) *domain.Task {
	return &domain.Task{
		ID:          d.ID.Hex(),
		UserID:      d.UserID,
		Title:       d.Title,
		Description: d.Description,
		Duration:    d.Duration,
		Priority:    d.Priority,
		IsHard:      d.IsHard,
		Status:      d.Status,
		Deadline:    d.Deadline,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func (r *taskMongoRepo) Create(ctx context.Context, task *domain.Task) error {
	doc := toDoc(task)
	doc.ID = primitive.NewObjectID()
	res, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	task.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *taskMongoRepo) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid task id")
	}
	var doc taskDocument
	if err := r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return fromDoc(doc), nil
}

func (r *taskMongoRepo) FindByUserID(ctx context.Context, userID string) ([]domain.Task, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	tasks := []domain.Task{}
	for cursor.Next(ctx) {
		var doc taskDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		tasks = append(tasks, *fromDoc(doc))
	}
	return tasks, cursor.Err()
}

func (r *taskMongoRepo) Update(ctx context.Context, id string, input domain.UpdateTaskInput) (*domain.Task, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid task id")
	}

	set := bson.M{"updated_at": time.Now()}
	if input.Title != nil {
		set["title"] = *input.Title
	}
	if input.Description != nil {
		set["description"] = *input.Description
	}
	if input.Duration != nil {
		set["duration"] = *input.Duration
	}
	if input.Priority != nil {
		set["priority"] = *input.Priority
	}
	if input.IsHard != nil {
		set["is_hard"] = *input.IsHard
	}
	if input.Status != nil {
		set["status"] = *input.Status
	}
	if input.Deadline != nil {
		set["deadline"] = *input.Deadline
	}

	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)
	var doc taskDocument
	err = r.collection.FindOneAndUpdate(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": set},
		opts,
	).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return fromDoc(doc), nil
}

func (r *taskMongoRepo) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid task id")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
