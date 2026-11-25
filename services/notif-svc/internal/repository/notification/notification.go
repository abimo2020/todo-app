package notification

import (
	"context"

	"github.com/abimo2020/notif-svc/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	coll *mongo.Collection
}

func New(db *mongo.Database) *Repository {
	return &Repository{coll: db.Collection("notifications")}
}

// membuat index untuk query daftar terbaru
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	})
	return err
}

func (r *Repository) Save(ctx context.Context, notif model.Notification) error {
	_, err := r.coll.InsertOne(ctx, notif)
	if mongo.IsDuplicateKeyError(err) {
		return model.ErrDuplicate
	}
	return err
}

func (r *Repository) List(ctx context.Context, limit int64) ([]model.Notification, error) {
	cur, err := r.coll.Find(ctx, bson.D{},
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetLimit(limit),
	)
	if err != nil {
		return nil, err
	}

	out := []model.Notification{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
