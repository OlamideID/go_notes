package notes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repo struct {
	call *mongo.Collection
}

func NewRepo(db *mongo.Database) *Repo {
	return &Repo{
		call: db.Collection("notes"),
	}
}

func (r *Repo) CreateNote(ctx context.Context, note *Note) (Note, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.call.InsertOne(opCtx, note)
	if err != nil {
		return Note{}, fmt.Errorf("failed to create note: %v", err)
	}
	return *note, nil
}

func (r *Repo) GetAllNotes(ctx context.Context) ([]Note, error) {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.call.Find(opCtx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to find notes: %v", err)
	}
	defer cursor.Close(opCtx)

	var notes []Note
	if err := cursor.All(opCtx, &notes); err != nil {
		return nil, fmt.Errorf("failed to decode notes: %v", err)
	}

	return notes, nil
}

func (r *Repo) List(ctx context.Context) ([]Note, error) {
	opctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{}

	cursor, err := r.call.Find(opctx, filter)
	if err != nil {
		fmt.Errorf("notes finding failed %w", err)
	}

	defer cursor.Close(opctx)

	var notes []Note

	if err := cursor.All(opctx, &notes); err != nil {
		return nil, fmt.Errorf("Wahala dey, no decoding: %w", err)
	}

	return notes, nil
}

func (r *Repo) GetNoteByID(ctx context.Context, id primitive.ObjectID) (*Note, error) {
	opctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	var note Note
	err := r.call.FindOne(opctx, filter, options.FindOne()).Decode(&note)
	if err != nil {
		return &Note{}, fmt.Errorf("failed to find note: %w", err)
	}
	return &note, nil
}

func (r *Repo) DeleteNotebyID(ctx context.Context, id primitive.ObjectID) error {
	opctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	res, err := r.call.DeleteOne(opctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *Repo) UpdateNote(ctx context.Context, id primitive.ObjectID, req UpdateNoteRequest) (*Note, error) {
	opctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"title":      req.Title,
			"content":    req.Content,
			"pinned":     req.Pinned,
			"updated_at": time.Now().UTC(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedNote Note
	err := r.call.FindOneAndUpdate(opctx, filter, update, opts).Decode(&updatedNote)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return &updatedNote, nil
}
