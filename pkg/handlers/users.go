package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/auth"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
)

type UsersRepoInterface interface {
	GetByID(string) (*user.User, error)
	Register(string, string) (*user.User, error)
	Authorize(string, string) (*user.User, error)
}

type SessionsManagerInterface interface {
	Create(string) (*session.Session, error)
}

type UsersHandler struct {
	//UsersRepo *user.UsersRepo
	UsersRepo UsersRepoInterface
	Logger    *zap.SugaredLogger
	Sessions  SessionsManagerInterface
}

type loginForm struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type tokenResponseData struct {
	Token string `json:"token"`
}

func (h *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {
	request, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	urd := &loginForm{}
	err = json.Unmarshal(request, urd)
	if err != nil {
		h.Logger.Errorf(`BadRequest. %s`, err.Error())
		http.Error(w, `BadRequest.Can'tParseRequestBody'`, http.StatusBadRequest)
		return
	}

	u, err := h.UsersRepo.Register(urd.Login, urd.Password)
	if err == user.ErrAlreadyRegistered {
		h.Logger.Errorf(`BadRequest. %s`, err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(
			[]byte(
				`{"errors":[{"location":"body","param":"username","value":"` + urd.Login + `","msg":"already exists"}]}`,
			),
		)
		return
	}
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	// Session
	sess, err := h.Sessions.Create(u.ID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}
	/* JWT entertainment */
	tokenString, err := auth.GenerateToken(u, sess.ID)

	tok := &tokenResponseData{Token: tokenString}

	resp, err := json.Marshal(tok)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (h *UsersHandler) Login(w http.ResponseWriter, r *http.Request) {
	request, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	urd := &loginForm{}
	err = json.Unmarshal(request, urd)
	if err != nil {
		h.Logger.Errorf(`BadRequest. %s`, err.Error())
		http.Error(w, `BadRequest.Can'tParseRequestBody'`, http.StatusBadRequest)
		return
	}

	u, err := h.UsersRepo.Authorize(urd.Login, urd.Password)
	if err == user.ErrNoUser {
		h.Logger.Errorf(`BadRequest. %s`, err.Error())
		jsonMessage(w, http.StatusBadRequest, "user not found")
		return
	}
	if err == user.ErrBadPass {
		h.Logger.Errorf(`BadRequest. %s`, err.Error())
		jsonMessage(w, http.StatusBadRequest, "invalid password")
		return
	}
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}

	// Session
	sess, err := h.Sessions.Create(u.ID)
	if err != nil {
		h.Logger.Errorf(`InternalServerError. %s`, err.Error())
		http.Error(w, `InternalServerError`, http.StatusInternalServerError)
		return
	}
	/* JWT entertainment */
	tokenString, err := auth.GenerateToken(u, sess.ID)

	tok := &tokenResponseData{Token: tokenString}

	resp, err := json.Marshal(tok)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func jsonMessage(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"message":"` + message + `"}`))
}
