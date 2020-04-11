package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http/httptest"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"testing"
	"time"
)

func TestHandlerGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := NewMockUsersRepoInterface(ctrl)
	postsRepo := NewMockPostsRepoInterface(ctrl)

	service := PostsHandler{
		UsersRepo: userRepo,
		PostsRepo: postsRepo,
		Logger:    zap.NewNop().Sugar(),
	}

	login := "login"
	uid := "userid"
	pid := "postid"
	cid := "commentid"

	resultUser := &user.User{
		Login: login,
		ID:    uid,
	}

	resultPost := &post.Post{
		Score:    0,
		Views:    0,
		Type:     "text",
		Title:    "title",
		Text:     "text",
		Author:   resultUser,
		Category: "programming",
		Votes: []post.Vote{
			{User: uid, Vote: 1},
		},
		Comments: []post.Comment{
			{
				Created: time.Now(),
				Author:  resultUser,
				Body:    "commentbody",
				ID:      cid,
			},
		},
		Created:          time.Now(),
		UpvotePercentage: 100,
		ID:               pid,
	}

	// good getall
	postsRepo.EXPECT().GetAll().Return([]*post.Post{resultPost}, nil)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	service.List(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img, _ := json.Marshal([]*post.Post{resultPost})
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// err getall
	postsRepo.EXPECT().GetAll().Return(nil, errors.New(""))

	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()

	service.List(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError.DatabaseError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), string(img))
		return
	}

	//////////////////////////////////////////////////
	// good category
	postsRepo.EXPECT().GetByCategory("programming").Return([]*post.Post{resultPost}, nil)

	req = httptest.NewRequest("GET", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"category_name": "programming",
	})
	w = httptest.NewRecorder()

	service.ListByCategory(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img, _ = json.Marshal([]*post.Post{resultPost})
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// err category
	postsRepo.EXPECT().GetByCategory("programming").Return(nil, errors.New(""))

	req = httptest.NewRequest("GET", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"category_name": "programming",
	})
	w = httptest.NewRecorder()

	service.ListByCategory(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError.DatabaseError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// err bad category
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()

	service.ListByCategory(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest.BadCategory`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	//////////////////////////////////////////
	// good get post
	postsRepo.EXPECT().GetByID(pid).Return(resultPost, nil)

	req = httptest.NewRequest("GET", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.GetPost(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img, _ = json.Marshal(resultPost)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// err category
	postsRepo.EXPECT().GetByID(pid).Return(nil, errors.New(""))

	req = httptest.NewRequest("GET", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.GetPost(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError.DatabaseError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// err bad category
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()

	service.GetPost(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest.BadPostID`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// good author
	postsRepo.EXPECT().GetByAuthor(login).Return([]*post.Post{resultPost}, nil)

	req = httptest.NewRequest("GET", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"user_login": login,
	})
	w = httptest.NewRecorder()

	service.ListByAuthor(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img, _ = json.Marshal([]*post.Post{resultPost})
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// err author
	postsRepo.EXPECT().GetByAuthor(login).Return(nil, errors.New(""))

	req = httptest.NewRequest("GET", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"user_login": login,
	})
	w = httptest.NewRecorder()

	service.ListByAuthor(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError.DatabaseError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// err bad author
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()

	service.ListByAuthor(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest.BadUserLogin`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}
}

func TestHandlerUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := NewMockUsersRepoInterface(ctrl)
	postsRepo := NewMockPostsRepoInterface(ctrl)

	service := PostsHandler{
		UsersRepo: userRepo,
		PostsRepo: postsRepo,
		Logger:    zap.NewNop().Sugar(),
	}

	login := "login"
	uid := "userid"
	pid := "postid"
	cid := "commentid"
	sid := "sessionid"
	//token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uX2lkIjoiaWRkIiwidXNlciI6eyJ1c2VybmFtZSI6ImxvZ2dnZ2dpbiIsImlkIjoiaWRkZGRkZCJ9fQ.AnBam3t75fSHoWx6lFnzp1MBW85ZQKf4ee5SshSiLHk"

	resultUser := &user.User{
		Login: login,
		ID:    uid,
	}

	resultPost := &post.Post{
		Score:    1,
		Views:    0,
		Type:     "text",
		Title:    "title",
		Text:     "text",
		Author:   resultUser,
		Category: "programming",
		Votes: []post.Vote{
			{User: uid, Vote: 1},
		},
		Created:          time.Now().Truncate(0),
		UpvotePercentage: 100,
		ID:               pid,
	}

	// good add
	userRepo.EXPECT().GetByID(uid).Return(resultUser, nil)
	postsRepo.EXPECT().Add(resultPost).Return(resultPost, nil)

	bts, _ := json.Marshal(resultPost)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	req = req.WithContext(context.WithValue(req.Context(), "session", &session.Session{
		ID:      sid,
		UserID:  uid,
		Expires: time.Now().Add(time.Hour),
	}))
	w := httptest.NewRecorder()

	service.Add(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img, _ := bts, 0
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// add bad json body
	req = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()

	service.Add(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// add. bad session
	bts, _ = json.Marshal(resultPost)
	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	w = httptest.NewRecorder()

	service.Add(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// add. repo err
	userRepo.EXPECT().GetByID(uid).Return(resultUser, nil)
	postsRepo.EXPECT().Add(resultPost).Return(nil, errors.New(""))

	bts, _ = json.Marshal(resultPost)
	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	req = req.WithContext(context.WithValue(req.Context(), "session", &session.Session{
		ID:      sid,
		UserID:  uid,
		Expires: time.Now().Add(time.Hour),
	}))
	w = httptest.NewRecorder()

	service.Add(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	///////////////////////////////////////////////
	// delete post. good
	postsRepo.EXPECT().DeletePost(pid).Return(nil)

	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.DeletePost(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`{"message":"success"}`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	//delete post. bad id
	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	w = httptest.NewRecorder()

	service.DeletePost(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest.BadPostID`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// delete post. repo err
	postsRepo.EXPECT().DeletePost(pid).Return(errors.New(""))

	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.DeletePost(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	/////////////////////////////////////////////////////////////////////////
	cbody := "commentbody"
	resultPost.Comments = []post.Comment{
		{
			Created: time.Now(),
			Author:  resultUser,
			Body:    cbody,
			ID:      cid,
		},
	}
	// add comment. good
	userRepo.EXPECT().GetByID(uid).Return(resultUser, nil)
	postsRepo.EXPECT().Comment(pid, resultUser, cbody).Return(resultPost, nil)

	bts, _ = json.Marshal(&comment{Comment: cbody})
	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	req = req.WithContext(context.WithValue(req.Context(), "session", &session.Session{
		ID:      sid,
		UserID:  uid,
		Expires: time.Now().Add(time.Hour),
	}))
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.Comment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	bts, _ = json.Marshal(resultPost)
	img, _ = bts, 0
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// add comment. repo comment err
	userRepo.EXPECT().GetByID(uid).Return(resultUser, nil)
	postsRepo.EXPECT().Comment(pid, resultUser, cbody).Return(nil, errors.New(""))

	bts, _ = json.Marshal(&comment{Comment: cbody})
	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	req = req.WithContext(context.WithValue(req.Context(), "session", &session.Session{
		ID:      sid,
		UserID:  uid,
		Expires: time.Now().Add(time.Hour),
	}))
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.Comment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// add comment. user repo err
	userRepo.EXPECT().GetByID(uid).Return(nil, errors.New(""))

	bts, _ = json.Marshal(&comment{Comment: cbody})
	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	req = req.WithContext(context.WithValue(req.Context(), "session", &session.Session{
		ID:      sid,
		UserID:  uid,
		Expires: time.Now().Add(time.Hour),
	}))
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.Comment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// add comment. sess err
	bts, _ = json.Marshal(&comment{Comment: cbody})
	req = httptest.NewRequest("POST", "/", bytes.NewReader(bts))
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.Comment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// add comment. body err
	req = httptest.NewRequest("POST", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.Comment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// add comment. bad post id
	req = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()

	service.Comment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest.BadPostID`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	/////////////////////////////////////
	resultPost.Comments = nil
	// delete comment. good
	postsRepo.EXPECT().DeleteComment(pid, cid).Return(resultPost, nil)

	req = httptest.NewRequest("POST", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"post_id":    pid,
		"comment_id": cid,
	})
	w = httptest.NewRecorder()

	service.DeleteComment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	bts, _ = json.Marshal(resultPost)
	img = bts
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// delete comment. repo delete err
	postsRepo.EXPECT().DeleteComment(pid, cid).Return(nil, errors.New(""))

	req = httptest.NewRequest("POST", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"post_id":    pid,
		"comment_id": cid,
	})
	w = httptest.NewRecorder()

	service.DeleteComment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// delete comment. bad vars
	req = httptest.NewRequest("POST", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.DeleteComment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest.BadCommentID`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}
	//
	req = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()

	service.DeleteComment(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest.BadPostID`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	////////////////////////////////////////////////////
	// vote. good
	postsRepo.EXPECT().Upvote(pid, uid).Return(resultPost, nil)
	postsRepo.EXPECT().Downvote(pid, uid).Return(resultPost, nil)
	postsRepo.EXPECT().Unvote(pid, uid).Return(resultPost, nil)

	req = httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), "session", &session.Session{
		ID:      sid,
		UserID:  uid,
		Expires: time.Now().Add(time.Hour),
	}))
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.Upvote(w, req)
	service.Downvote(w, req)
	service.Unvote(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	bts, _ = json.Marshal(resultPost)
	img, _ = bts, 0
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// vote. repo err
	postsRepo.EXPECT().Upvote(pid, uid).Return(nil, errors.New(""))
	postsRepo.EXPECT().Downvote(pid, uid).Return(nil, errors.New(""))
	postsRepo.EXPECT().Unvote(pid, uid).Return(nil, errors.New(""))

	req = httptest.NewRequest("POST", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), "session", &session.Session{
		ID:      sid,
		UserID:  uid,
		Expires: time.Now().Add(time.Hour),
	}))
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.Upvote(w, req)
	service.Downvote(w, req)
	service.Unvote(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// vote. repo err
	req = httptest.NewRequest("POST", "/", nil)
	req = mux.SetURLVars(req, map[string]string{
		"post_id": pid,
	})
	w = httptest.NewRecorder()

	service.Upvote(w, req)
	service.Downvote(w, req)
	service.Unvote(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`InternalServerError`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}

	// vote. bad vars
	req = httptest.NewRequest("POST", "/", nil)
	w = httptest.NewRecorder()

	service.Upvote(w, req)
	service.Downvote(w, req)
	service.Unvote(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = []byte(`BadRequest.BadPostID`)
	if !bytes.Contains(body, img) {
		t.Errorf("Invalid.\n%s\n%s", string(body), img)
		return
	}
}
