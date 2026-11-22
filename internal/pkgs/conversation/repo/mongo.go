package repo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cgcodeboy/sim-studio/internal/pkgs/conversation/domain"
)

// Mongo 把会话和工程级事实放在数据库里。
type Mongo struct {
	sessions *mongo.Collection
	facts    *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{sessions: db.Collection("sessions"), facts: db.Collection("facts")}
}

func (r *Mongo) Load(ctx context.Context, id string) (domain.Session, error) {
	var session domain.Session
	err := r.sessions.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Session{ID: id}, nil
	}
	return session, err
}

func (r *Mongo) Save(ctx context.Context, session domain.Session) error {
	_, err := r.sessions.ReplaceOne(ctx, bson.M{"_id": session.ID}, session, replaceUpsert())
	return err
}

func (r *Mongo) List(ctx context.Context, project string) ([]domain.Session, error) {
	cursor, err := r.sessions.Find(ctx, bson.M{"project": project})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []domain.Session
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Mongo) Delete(ctx context.Context, id string) error {
	_, err := r.sessions.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *Mongo) Facts(ctx context.Context, project string) ([]domain.Fact, error) {
	var doc struct {
		Facts []domain.Fact `bson:"facts"`
	}
	err := r.facts.FindOne(ctx, bson.M{"_id": project}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return doc.Facts, err
}

func (r *Mongo) ReplaceFacts(ctx context.Context, project string, facts []domain.Fact) error {
	_, err := r.facts.ReplaceOne(ctx, bson.M{"_id": project}, bson.M{"_id": project, "facts": facts}, replaceUpsert())
	return err
}
