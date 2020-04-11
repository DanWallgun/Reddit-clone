package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
)

type PostsRepoInterface interface {
	GetAll() ([]*post.Post, error)
	GetByCategory(string) ([]*post.Post, error)
	GetByID(string) (*post.Post, error)
	GetByAuthor(string) ([]*post.Post, error)
	Add(*post.Post) (*post.Post, error)
	Comment(string, *user.User, string) (*post.Post, error)
	DeleteComment(string, string) (*post.Post, error)
	DeletePost(string) error
	Upvote(string, string) (*post.Post, error)
	Downvote(string, string) (*post.Post, error)
	Unvote(string, string) (*post.Post, error)
}

type PostsHandler struct {
	UsersRepo UsersRepoInterface
	PostsRepo PostsRepoInterface
	Logger    *zap.SugaredLogger
}

type comment struct {
	Comment string `json:"comment"`
}

func (h *PostsHandler) List(w http.ResponseWriter, r *http.Request) {
	posts, err := h.PostsRepo.GetAll()
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError.DatabaseError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(posts)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category, ok := vars["category_name"]
	if !ok {
		http.Error(w, `BadRequest.BadCategory`, http.StatusBadRequest)
		return
	}

	posts, err := h.PostsRepo.GetByCategory(category)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError.DatabaseError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(posts)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars["post_id"]
	if !ok {
		http.Error(w, `BadRequest.BadPostID`, http.StatusBadRequest)
		return
	}

	p, err := h.PostsRepo.GetByID(postID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError.DatabaseError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) ListByAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userLogin, ok := vars["user_login"]
	if !ok {
		http.Error(w, `BadRequest.BadUserLogin`, http.StatusBadRequest)
		return
	}

	posts, err := h.PostsRepo.GetByAuthor(userLogin)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError.DatabaseError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(posts)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) Add(w http.ResponseWriter, r *http.Request) {
	p := new(post.Post)
	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	err = json.Unmarshal(data, p)
	if err != nil {
		h.Logger.Errorf(`BadRequest. %s`, err.Error())
		http.Error(w, `BadRequest`, http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	userID := sess.UserID
	u, err := h.UsersRepo.GetByID(userID)
	p.Author = u

	p, err = h.PostsRepo.Add(p)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) Comment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars["post_id"]
	if !ok {
		http.Error(w, `BadRequest.BadPostID`, http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	c := &comment{}
	err = json.Unmarshal(data, c)
	if err != nil {
		h.Logger.Errorf(`BadRequest. %s`, err.Error())
		http.Error(w, `BadRequest`, http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	userID := sess.UserID
	u, err := h.UsersRepo.GetByID(userID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	p, err := h.PostsRepo.Comment(postID, u, c.Comment)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars["post_id"]
	if !ok {
		http.Error(w, `BadRequest.BadPostID`, http.StatusBadRequest)
		return
	}
	commentID, ok := vars["comment_id"]
	if !ok {
		http.Error(w, `BadRequest.BadCommentID`, http.StatusBadRequest)
		return
	}

	p, err := h.PostsRepo.DeleteComment(postID, commentID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars["post_id"]
	if !ok {
		http.Error(w, `BadRequest.BadPostID`, http.StatusBadRequest)
		return
	}

	err := h.PostsRepo.DeletePost(postID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	jsonMessage(w, http.StatusOK, "success")
}

func (h *PostsHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars["post_id"]
	if !ok {
		http.Error(w, `BadRequest.BadPostID`, http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	userID := sess.UserID

	p, err := h.PostsRepo.Upvote(postID, userID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars["post_id"]
	if !ok {
		http.Error(w, `BadRequest.BadPostID`, http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	userID := sess.UserID

	p, err := h.PostsRepo.Downvote(postID, userID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *PostsHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars["post_id"]
	if !ok {
		http.Error(w, `BadRequest.BadPostID`, http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	userID := sess.UserID

	p, err := h.PostsRepo.Unvote(postID, userID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
