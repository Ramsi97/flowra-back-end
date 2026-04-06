package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
	"github.com/Ramsi97/flowra-back-end/internal/schedule/repository/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type scheduleDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    string             `bson:"user_id"`
	TaskID    string             `bson:"task_id"`
	Title     string             `bson:"title"`
	StartTime time.Time          `bson:"start_time"`
	EndTime   time.Time          `bson:"end_time"`
	IsHard    bool               `bson:"is_hard"`
	Status    string             `bson:"status"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type scheduleMongoRepo struct {
	col *mongo.Collection
}

func NewScheduleMongoRepo(db *mongo.Database) interfaces.ScheduleRepository {
	return &scheduleMongoRepo{col: db.Collection("schedule_items")}
}

func fromDoc(d scheduleDocument) domain.ScheduleItem {
	return domain.ScheduleItem{
		ID:        d.ID.Hex(),
		UserID:    d.UserID,
		TaskID:    d.TaskID,
		Title:     d.Title,
		StartTime: d.StartTime,
		EndTime:   d.EndTime,
		IsHard:    d.IsHard,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func toOID(id string) (primitive.ObjectID, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid schedule item id")
	}
	return oid, nil
}

func (r *scheduleMongoRepo) InsertMany(ctx context.Context, items []domain.ScheduleItem) error {
	if len(items) == 0 {
		return nil
	}
	docs := make([]interface{}, 0, len(items))
	oids := make([]primitive.ObjectID, len(items))
	for i := range items {
		oid := primitive.NewObjectID()
		oids[i] = oid
		docs = append(docs, scheduleDocument{
			ID:        oid,
			UserID:    items[i].UserID,
			TaskID:    items[i].TaskID,
			Title:     items[i].Title,
			StartTime: items[i].StartTime,
			EndTime:   items[i].EndTime,
			IsHard:    items[i].IsHard,
			Status:    items[i].Status,
			CreatedAt: items[i].CreatedAt,
			UpdatedAt: items[i].UpdatedAt,
		})
	}
	if _, err := r.col.InsertMany(ctx, docs); err != nil {
		return err
	}
	for i := range items {
		items[i].ID = oids[i].Hex()
	}
	return nil
}

func (r *scheduleMongoRepo) FindByID(ctx context.Context, id string) (*domain.ScheduleItem, error) {
	oid, err := toOID(id)
	if err != nil {
		return nil, err
	}
	var doc scheduleDocument
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	item := fromDoc(doc)
	return &item, nil
}

func dayRange(date time.Time) (time.Time, time.Time) {
	y, m, d := date.Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return start, start.Add(24 * time.Hour)
}

func (r *scheduleMongoRepo) FindByUserAndDate(ctx context.Context, userID string, date time.Time) ([]domain.ScheduleItem, error) {
	start, end := dayRange(date)
	return r.FindByUserAndDateRange(ctx, userID, start, end)
}

func (r *scheduleMongoRepo) FindByUserAndDateRange(ctx context.Context, userID string, start, end time.Time) ([]domain.ScheduleItem, error) {
	cursor, err := r.col.Find(ctx, bson.M{
		"user_id":    userID,
		"start_time": bson.M{"$gte": start, "$lt": end},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []domain.ScheduleItem
	for cursor.Next(ctx) {
		var doc scheduleDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		items = append(items, fromDoc(doc))
	}
	return items, cursor.Err()
}

func (r *scheduleMongoRepo) UpdateItemTimes(ctx context.Context, id string, start, end time.Time) error {
	oid, err := toOID(id)
	if err != nil {
		return err
	}
	_, err = r.col.UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"start_time": start, "end_time": end, "updated_at": time.Now()}},
	)
	return err
}

func (r *scheduleMongoRepo) Update(ctx context.Context, id string, input domain.UpdateItemInput) (*domain.ScheduleItem, error) {
	oid, err := toOID(id)
	if err != nil {
		return nil, err
	}
	set := bson.M{"updated_at": time.Now()}
	if input.StartTime != nil {
		set["start_time"] = *input.StartTime
	}
	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)
	var doc scheduleDocument
	if err := r.col.FindOneAndUpdate(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": set},
		opts,
	).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	item := fromDoc(doc)
	return &item, nil
}

func (r *scheduleMongoRepo) DeleteByID(ctx context.Context, id string) error {
	oid, err := toOID(id)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (r *scheduleMongoRepo) DeleteByUserAndDate(ctx context.Context, userID string, date time.Time) error {
	start, end := dayRange(date)
	return r.DeleteByUserAndDateRange(ctx, userID, start, end)
}

func (r *scheduleMongoRepo) DeleteByUserAndDateRange(ctx context.Context, userID string, start, end time.Time) error {
	_, err := r.col.DeleteMany(ctx, bson.M{
		"user_id":    userID,
		"start_time": bson.M{"$gte": start, "$lt": end},
	})
	return err
}

func (r *scheduleMongoRepo) DeleteByUserAndMonth(ctx context.Context, userID string, year, month int) error {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return r.DeleteByUserAndDateRange(ctx, userID, start, start.AddDate(0, 1, 0))
}
