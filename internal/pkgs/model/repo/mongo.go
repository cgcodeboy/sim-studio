package repo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
)

type Mongo struct {
	projects *mongo.Collection
}

func New(db *mongo.Database) *Mongo {
	return &Mongo{projects: db.Collection("projects")}
}

func (r *Mongo) List(ctx context.Context) ([]domain.Summary, error) {
	cur, err := r.projects.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1, "name": 1}))
	if err != nil {
		return nil, err
	}
	var out []domain.Summary
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []domain.Summary{}
	}
	return out, nil
}

func (r *Mongo) IDs(ctx context.Context) ([]string, error) {
	cur, err := r.projects.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	var ids []string
	for cur.Next(ctx) {
		var row struct {
			ID string `bson:"_id"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		ids = append(ids, row.ID)
	}
	if err := cur.Close(ctx); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *Mongo) Read(ctx context.Context, id string) (domain.Project, error) {
	var p domain.Project
	err := r.projects.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Project{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, err
	}
	domain.Normalize(&p)
	return p, nil
}

func (r *Mongo) Insert(ctx context.Context, p domain.Project) error {
	_, err := r.projects.InsertOne(ctx, p)
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrDuplicate
	}
	return err
}

func (r *Mongo) Replace(ctx context.Context, expected int64, p domain.Project) (bool, error) {
	res, err := r.projects.ReplaceOne(ctx, bson.M{"_id": p.ID, "revision": expected}, p)
	if err != nil {
		return false, err
	}
	return res.MatchedCount == 1, nil
}

func (r *Mongo) Delete(ctx context.Context, id string) (bool, error) {
	res, err := r.projects.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return res.DeletedCount == 1, nil
}
