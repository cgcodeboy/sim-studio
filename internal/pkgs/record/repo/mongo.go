package repo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
)

type Mongo struct {
	records *mongo.Collection
}

func New(db *mongo.Database) *Mongo {
	return &Mongo{records: db.Collection("records")}
}

func (r *Mongo) Ensure(ctx context.Context) error {
	_, err := r.records.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "requestId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "projectId", Value: 1}}},
		{
			Keys: bson.D{{Key: "active", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("active_one").
				SetPartialFilterExpression(bson.M{"active": int32(1)}),
		},
	})
	if err != nil {
		return err
	}
	_, _ = r.records.Indexes().DropOne(ctx, "status_1")
	return nil
}

func (r *Mongo) Insert(ctx context.Context, record domain.Record) error {
	_, err := r.records.InsertOne(ctx, record)
	if !mongo.IsDuplicateKeyError(err) {
		return err
	}
	_, findErr := r.FindByRequestID(ctx, record.RequestID)
	if findErr == nil {
		return domain.ErrDuplicate
	}
	return domain.ErrActive
}

func (r *Mongo) FindByRequestID(ctx context.Context, requestID string) (domain.Record, error) {
	var record domain.Record
	err := r.records.FindOne(ctx, bson.M{"requestId": requestID}).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Record{}, domain.ErrNotFound
	}
	return record, err
}

func (r *Mongo) Read(ctx context.Context, id string) (domain.Record, error) {
	var record domain.Record
	err := r.records.FindOne(ctx, bson.M{"_id": id}).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Record{}, domain.ErrNotFound
	}
	return record, err
}

func (r *Mongo) MatchRevision(ctx context.Context, id string, revision int64, next domain.Record) (int64, error) {
	set := bson.M{
		"status": next.Status, "speed": next.Speed, "source": next.Source,
		"revision": revision + 1, "updatedAt": sampleTime(),
		"controlRequestId": next.ControlRequestID, "controlOp": next.ControlOp,
	}
	update := bson.M{"$set": set}
	if next.Sample != nil {
		set["sample"] = next.Sample
	}
	if next.Status == "finished" || next.Status == "interrupted" {
		update["$unset"] = bson.M{"active": ""}
	} else if next.Active != nil {
		set["active"] = *next.Active
	}
	res, err := r.records.UpdateOne(ctx, bson.M{"_id": id, "revision": revision}, update)
	if err != nil {
		return 0, err
	}
	return res.MatchedCount, nil
}

func (r *Mongo) CountActive(ctx context.Context, projectID string) (int64, error) {
	return r.records.CountDocuments(ctx, bson.M{
		"projectId": projectID,
		"status":    bson.M{"$in": []string{"running", "paused"}},
	})
}

func (r *Mongo) DeleteByProject(ctx context.Context, projectID string) error {
	n, err := r.CountActive(ctx, projectID)
	if err != nil {
		return err
	}
	if n > 0 {
		return domain.ErrActive
	}
	_, err = r.records.DeleteMany(ctx, bson.M{"projectId": projectID})
	return err
}

func (r *Mongo) InterruptActive(ctx context.Context) error {
	_, err := r.records.UpdateMany(ctx,
		bson.M{"status": bson.M{"$in": []string{"running", "paused"}}},
		bson.M{
			"$set":   bson.M{"status": "interrupted", "updatedAt": sampleTime()},
			"$unset": bson.M{"active": ""},
		},
	)
	return err
}

func (r *Mongo) DeleteOrphans(ctx context.Context, projectIDs []string) error {
	_, err := r.records.DeleteMany(ctx, bson.M{"projectId": bson.M{"$nin": projectIDs}})
	return err
}
