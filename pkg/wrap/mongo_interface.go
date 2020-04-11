package wrap

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionHelper interface {
	Find(context.Context, interface{}, ...*options.FindOptions) (CursorHelper, error)
	InsertOne(context.Context, interface{}, ...*options.InsertOneOptions) (InsertOneResultHelper, error)
	DeleteOne(context.Context, interface{}, ...*options.DeleteOptions) (DeleteResultHelper, error)
	ReplaceOne(context.Context, interface{}, interface{}, ...*options.ReplaceOptions) (UpdateResultHelper, error)
}

type CursorHelper interface {
	Close(context.Context) error
	Next(context.Context) bool
	Decode(interface{}) error
	Err() error
}
type InsertOneResultHelper interface{}
type DeleteResultHelper interface{}
type UpdateResultHelper interface{}

type mongoCollection struct {
	coll *mongo.Collection
}
type mongoCursor struct {
	cur *mongo.Cursor
}
type mongoInsertOneResult struct {
	res *mongo.InsertOneResult
}
type mongoDeleteResult struct {
	res *mongo.DeleteResult
}
type mongoUpdateResult struct {
	res *mongo.UpdateResult
}

func (mc *mongoCollection) Find(
	ctx context.Context,
	filter interface{},
	opts ...*options.FindOptions,
) (CursorHelper, error) {
	cur, err := mc.coll.Find(ctx, filter, opts...)
	return &mongoCursor{cur: cur}, err
}

func (mc *mongoCollection) InsertOne(
	ctx context.Context,
	filter interface{},
	opts ...*options.InsertOneOptions,
) (InsertOneResultHelper, error) {
	res, err := mc.coll.InsertOne(ctx, filter, opts...)
	return &mongoInsertOneResult{res: res}, err
}

func (mc *mongoCollection) DeleteOne(
	ctx context.Context,
	filter interface{},
	opts ...*options.DeleteOptions,
) (DeleteResultHelper, error) {
	res, err := mc.coll.DeleteOne(ctx, filter, opts...)
	return &mongoDeleteResult{res: res}, err
}

func (mc *mongoCollection) ReplaceOne(
	ctx context.Context,
	filter interface{},
	replacement interface{},
	opts ...*options.ReplaceOptions,
) (UpdateResultHelper, error) {
	res, err := mc.coll.ReplaceOne(ctx, filter, replacement, opts...)
	return &mongoUpdateResult{res: res}, err
}

func (mc *mongoCursor) Close(ctx context.Context) error {
	return mc.cur.Close(ctx)
}

func (mc *mongoCursor) Next(ctx context.Context) bool {
	return mc.cur.Next(ctx)
}

func (mc *mongoCursor) Decode(val interface{}) error {
	return mc.cur.Decode(val)
}

func (mc *mongoCursor) Err() error {
	return mc.cur.Err()
}
