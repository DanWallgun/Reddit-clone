package post_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"redditclone/pkg/mocks"
	"redditclone/pkg/post"
	"redditclone/pkg/user"
	"redditclone/pkg/wrap"
	"reflect"
	"testing"
	"time"
)

//go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html

/* Cursor that returns no error but given post */
func coolCursor(pt *post.Post) wrap.CursorHelper {
	var curHr wrap.CursorHelper
	curHr = &mocks.CursorHelper{}
	curHr.(*mocks.CursorHelper).
		On("Close", context.Background()).
		Return(nil)
	curHr.(*mocks.CursorHelper).
		On("Next", context.Background()).
		Return(true).Once()
	curHr.(*mocks.CursorHelper).
		On("Next", context.Background()).
		Return(false).Once()
	curHr.(*mocks.CursorHelper).
		On("Close", context.Background()).
		Return(nil)
	curHr.(*mocks.CursorHelper).
		On("Decode", &post.Post{}).
		Return(nil).Run(func(args mock.Arguments) {
		p := args.Get(0).(*post.Post)
		p.Score = pt.Score
		p.Views = pt.Views
		p.Type = pt.Type
		p.Category = pt.Category
		p.ID = pt.ID
		p.Author = pt.Author
		p.Comments = pt.Comments
		p.Votes = pt.Votes
	})
	curHr.(*mocks.CursorHelper).
		On("Err").
		Return(nil)
	return curHr
}

/* Cursor that returns error by bad decoding */
func badDecodeCursor() wrap.CursorHelper {
	var curHr wrap.CursorHelper
	curHr = &mocks.CursorHelper{}
	curHr.(*mocks.CursorHelper).
		On("Close", context.Background()).
		Return(nil)
	curHr.(*mocks.CursorHelper).
		On("Next", context.Background()).
		Return(true).Once()
	curHr.(*mocks.CursorHelper).
		On("Next", context.Background()).
		Return(false).Once()
	curHr.(*mocks.CursorHelper).
		On("Decode", &post.Post{}).
		Return(errors.New(""))
	curHr.(*mocks.CursorHelper).
		On("Err").
		Return(nil)
	return curHr
}

/* Cursor that returns error by Err() method */
func errCursor(pt *post.Post) wrap.CursorHelper {
	var curHr wrap.CursorHelper
	curHr = &mocks.CursorHelper{}
	curHr.(*mocks.CursorHelper).
		On("Close", context.Background()).
		Return(nil)
	curHr.(*mocks.CursorHelper).
		On("Next", context.Background()).
		Return(true).Once()
	curHr.(*mocks.CursorHelper).
		On("Next", context.Background()).
		Return(false).Once()
	curHr.(*mocks.CursorHelper).
		On("Decode", &post.Post{}).
		Return(nil).Run(func(args mock.Arguments) {
		p := args.Get(0).(*post.Post)
		p.Score = pt.Score
		p.Views = pt.Views
		p.Type = pt.Type
		p.Category = pt.Category
		p.ID = pt.ID
		p.Author = pt.Author
		p.Comments = pt.Comments
	})
	curHr.(*mocks.CursorHelper).
		On("Err").
		Return(errors.New(""))
	return curHr
}

func TestPostGet(t *testing.T) {
	var colHr wrap.CollectionHelper
	var curHr wrap.CursorHelper
	colHr = &mocks.CollectionHelper{}

	pt := &post.Post{
		ID:       "postid",
		Type:     "text",
		Category: "programming",
		Author: &user.User{
			Login: "userlogin",
			ID:    "userid",
		},
	}

	service := post.NewRepo(colHr)

	// 1
	curHr = coolCursor(pt)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{}).
		Return(curHr, nil).Once()

	ps, err := service.GetAll()
	if err != nil || !reflect.DeepEqual(ps[0], pt) {
		t.Errorf("Invalid.\n%#v\n%#v\nErr:%#v", ps[0], pt, err)
	}

	// 2
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"category": "programming"}).
		Return(nil, errors.New("")).Once()

	ps, err = service.GetByCategory("programming")
	if err == nil || ps != nil {
		t.Errorf("Invalid.%#v%#v", ps, err)
	}

	// 3
	curHr = badDecodeCursor()
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id": "postid"}).
		Return(curHr, nil).Once()

	pst, err := service.GetByID("postid")
	if err == nil || pst != nil {
		t.Errorf("Invalid.%#v%#v", pst, err)
	}

	// 4
	curHr = errCursor(pt)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"author.login": "userlogin"}).
		Return(curHr, nil).Once()

	ps, err = service.GetByAuthor("userlogin")
	if err == nil {
		t.Errorf("Invalid.\n%#v\n%#v", ps, err)
	}

	// 5
	curHr = coolCursor(pt)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id": "postid"}).
		Return(curHr, nil).Once()

	pst, err = service.GetByID("postid")
	if err != nil || !reflect.DeepEqual(pst, pt) {
		t.Errorf("Invalid.\n%#v\n%#v", pst, err)
	}
}

func TestPostPost(t *testing.T) {
	var colHr wrap.CollectionHelper
	var iorHr wrap.InsertOneResultHelper
	var drHr wrap.DeleteResultHelper
	colHr = &mocks.CollectionHelper{}
	iorHr = &mocks.InsertOneResultHelper{}
	drHr = &mocks.DeleteResultHelper{}

	ppost := &post.Post{
		Type:     "text",
		Category: "programming",
		Author: &user.User{
			Login: "userlogin",
			ID:    "userid",
		},
	}

	service := post.NewRepo(colHr)

	// POST
	// 1
	colHr.(*mocks.CollectionHelper).
		On("InsertOne", context.Background(), ppost).
		Return(iorHr, nil).Once()

	retPost, err := service.Add(ppost)
	if err != nil {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// 2
	colHr.(*mocks.CollectionHelper).
		On("InsertOne", context.Background(), ppost).
		Return(nil, errors.New("")).Once()

	retPost, err = service.Add(ppost)
	if err == nil || retPost != nil {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// DELETE POST
	// 1
	colHr.(*mocks.CollectionHelper).
		On("DeleteOne", context.Background(), bson.M{"id":"postid"}).
		Return(drHr, nil).Once()

	err = service.DeletePost("postid")
	if err != nil {
		t.Errorf("Invalid. Err:%#v", err)
	}

	// 2
	colHr.(*mocks.CollectionHelper).
		On("DeleteOne", context.Background(), bson.M{"id":"postid"}).
		Return(nil, errors.New("")).Once()

	err = service.DeletePost("postid")
	if err == nil {
		t.Errorf("Invalid. Err:%#v", err)
	}
}

func TestPostComment(t *testing.T) {
	var colHr wrap.CollectionHelper
	var urHr wrap.UpdateResultHelper
	colHr = &mocks.CollectionHelper{}
	urHr = &mocks.UpdateResultHelper{}

	u := &user.User{
		Login: "userlogin",
		ID:    "userid",
	}
	ppost := &post.Post{
		ID: "postid",
		Type:     "text",
		Category: "programming",
		Author: u,
	}

	service := post.NewRepo(colHr)
	//utils.IsRandom = false
	//defer func() {utils.IsRandom = true}()

	// COMMENT
	// 1
	curHr := coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(curHr, nil).Once()
	colHr.(*mocks.CollectionHelper).
		On("ReplaceOne", context.Background(), bson.M{"id":"postid"}, mock.Anything).
		Return(urHr, nil).Once()

	retPost, err := service.Comment("postid", u, "commentbody")
	if err != nil || retPost == nil || len(retPost.Comments) == 0 {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// 2
	curHr = coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(curHr, nil).Once()
	colHr.(*mocks.CollectionHelper).
		On("ReplaceOne", context.Background(), bson.M{"id":"postid"}, mock.Anything).
		Return(nil, errors.New("")).Once()

	retPost, err = service.Comment("postid", u, "commentbody")
	if err == nil || retPost != nil {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// 3
	curHr = coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(nil, errors.New("")).Once()
	retPost, err = service.Comment("postid", u, "commentbody")
	if err == nil || retPost != nil {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// DELETE COMMENT
	ppost.Comments = append(ppost.Comments, post.Comment{
		Created: time.Now(),
		Author:  u,
		Body:    "commentbody",
		ID:      "commentid",
	})
	// 1
	curHr = coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(curHr, nil).Once()
	colHr.(*mocks.CollectionHelper).
		On("ReplaceOne", context.Background(), bson.M{"id":"postid"}, mock.Anything).
		Return(urHr, nil).Once()

	retPost, err = service.DeleteComment("postid", "commentid")
	if err != nil || retPost == nil || len(retPost.Comments) > 0 {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// 2
	curHr = coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(curHr, nil).Once()
	colHr.(*mocks.CollectionHelper).
		On("ReplaceOne", context.Background(), bson.M{"id":"postid"}, mock.Anything).
		Return(nil, errors.New("")).Once()

	retPost, err = service.DeleteComment("postid", "commentid")
	if err == nil || retPost != nil {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// 3
	curHr = coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(nil, errors.New("")).Once()

	retPost, err = service.DeleteComment("postid", "commentid")
	if err == nil || retPost != nil {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}
}

func TestPostVote(t *testing.T) {
	var colHr wrap.CollectionHelper
	var urHr wrap.UpdateResultHelper
	colHr = &mocks.CollectionHelper{}
	urHr = &mocks.UpdateResultHelper{}

	u := &user.User{
		Login: "userlogin",
		ID:    "userid",
	}
	ppost := &post.Post{
		ID: "postid",
		Type:     "text",
		Category: "programming",
		Author: u,
	}

	service := post.NewRepo(colHr)

	// VOTING
	// 1
	curHr := coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(curHr, nil).Times(1)
	colHr.(*mocks.CollectionHelper).
		On("ReplaceOne", context.Background(), bson.M{"id":"postid"}, mock.Anything).
		Return(urHr, nil).Times(1)

	retPost, err := service.Upvote("postid", u.ID)
	if err != nil || retPost == nil || len(retPost.Votes) == 0 || retPost.Votes[0].User != u.ID {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}
	// 2
	curHr = coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(curHr, nil).Times(1)
	colHr.(*mocks.CollectionHelper).
		On("ReplaceOne", context.Background(), bson.M{"id":"postid"}, mock.Anything).
		Return(nil, errors.New("")).Times(1)

	retPost, err = service.Downvote("postid", u.ID)
	if err == nil || retPost != nil {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// 3
	curHr = coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(nil, errors.New("")).Times(1)

	retPost, err = service.Unvote("postid", u.ID)
	if err == nil || retPost != nil {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}

	// 4 REVOTE
	ppost.Votes = append(ppost.Votes, post.Vote{
		User: u.ID,
		Vote: -1,
	})
	curHr = coolCursor(ppost)
	colHr.(*mocks.CollectionHelper).
		On("Find", context.Background(), bson.M{"id":"postid"}).
		Return(curHr, nil).Times(1)
	colHr.(*mocks.CollectionHelper).
		On("ReplaceOne", context.Background(), bson.M{"id":"postid"}, mock.Anything).
		Return(urHr, nil).Times(1)

	retPost, err = service.Upvote("postid", u.ID)
	if err != nil || retPost == nil || len(retPost.Votes) == 0 || retPost.Votes[0].Vote != 1 {
		t.Errorf("Invalid.\n%#v\nErr:%#v", retPost, err)
	}
}