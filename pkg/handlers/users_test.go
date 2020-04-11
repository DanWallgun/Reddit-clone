package handlers

import (
	"bytes"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http/httptest"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"strings"
	"testing"
	"time"
)

func TestHandlerRegister(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sessRepo := NewMockSessionsManagerInterface(ctrl)
	userRepo := NewMockUsersRepoInterface(ctrl)
	service := UsersHandler{
		UsersRepo: userRepo,
		Logger:    zap.NewNop().Sugar(),
		Sessions:  sessRepo,
	}

	login := "logggggin"
	uid := "idddddd"
	sid := "idd"
	pass := "pasdasd"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uX2lkIjoiaWRkIiwidXNlciI6eyJ1c2VybmFtZSI6ImxvZ2dnZ2dpbiIsImlkIjoiaWRkZGRkZCJ9fQ.AnBam3t75fSHoWx6lFnzp1MBW85ZQKf4ee5SshSiLHk"

	resultUser := &user.User{
		Login: login, ID: uid,
	}
	resultSess := &session.Session{
		ID: sid, UserID: uid, Expires: time.Now(),
	}

	userRepo.EXPECT().Register(login, pass).Return(resultUser, nil)
	sessRepo.EXPECT().Create(uid).Return(resultSess, nil)

	req := httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w := httptest.NewRecorder()

	service.Register(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img := `{"token":"` + token + `"}`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	/////////////////////////////////////////////////////////////////////////
	// No body
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()

	service.Register(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `BadRequest.Can'tParseRequestBody'`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	///////////////////////////////////////
	// register repo error already registered
	userRepo.EXPECT().Register(login, pass).Return(nil, user.ErrAlreadyRegistered)
	req = httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w = httptest.NewRecorder()

	service.Register(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `{"errors":[{"location":"body","param":"username","value":"` + login + `","msg":"already exists"}]}`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	///////////////////////////////////////
	// register repo error
	userRepo.EXPECT().Register(login, pass).Return(nil, errors.New(""))
	req = httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w = httptest.NewRecorder()

	service.Register(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `InternalServerError`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	///////////////////////////////////////
	// register repo error
	userRepo.EXPECT().Register(login, pass).Return(resultUser, nil)
	sessRepo.EXPECT().Create(uid).Return(nil, errors.New(""))
	req = httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w = httptest.NewRecorder()

	service.Register(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `InternalServerError`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}
}

func TestHandlerLogin(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sessRepo := NewMockSessionsManagerInterface(ctrl)
	userRepo := NewMockUsersRepoInterface(ctrl)
	service := UsersHandler{
		UsersRepo: userRepo,
		Logger:    zap.NewNop().Sugar(),
		Sessions:  sessRepo,
	}

	login := "logggggin"
	uid := "idddddd"
	sid := "idd"
	pass := "pasdasd"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzZXNzaW9uX2lkIjoiaWRkIiwidXNlciI6eyJ1c2VybmFtZSI6ImxvZ2dnZ2dpbiIsImlkIjoiaWRkZGRkZCJ9fQ.AnBam3t75fSHoWx6lFnzp1MBW85ZQKf4ee5SshSiLHk"

	resultUser := &user.User{
		Login: login, ID: uid,
	}
	resultSess := &session.Session{
		ID: sid, UserID: uid, Expires: time.Now(),
	}

	userRepo.EXPECT().Authorize(login, pass).Return(resultUser, nil)
	sessRepo.EXPECT().Create(uid).Return(resultSess, nil)

	req := httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w := httptest.NewRecorder()

	service.Login(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img := `{"token":"` + token + `"}`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	/////////////////////////////////////////////////////////////////////////
	// No body
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()

	service.Login(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `BadRequest.Can'tParseRequestBody'`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	///////////////////////////////////////
	// login repo error
	userRepo.EXPECT().Authorize(login, pass).Return(nil, user.ErrNoUser)
	req = httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w = httptest.NewRecorder()

	service.Login(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `{"message":"user not found"}`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	///////////////////////////////////////
	// login repo error
	userRepo.EXPECT().Authorize(login, pass).Return(nil, user.ErrBadPass)
	req = httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w = httptest.NewRecorder()

	service.Login(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `{"message":"invalid password"}`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	///////////////////////////////////////
	// login repo error
	userRepo.EXPECT().Authorize(login, pass).Return(nil, errors.New(""))
	req = httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w = httptest.NewRecorder()

	service.Login(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `InternalServerError`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}

	///////////////////////////////////////
	// login sess repo error
	userRepo.EXPECT().Authorize(login, pass).Return(resultUser, nil)
	sessRepo.EXPECT().Create(uid).Return(nil, errors.New(""))
	req = httptest.NewRequest("GET", "/",
		strings.NewReader(`{"username":"`+login+`", "password": "`+pass+`"}`))
	w = httptest.NewRecorder()

	service.Login(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	img = `InternalServerError`
	if !bytes.Contains(body, []byte(img)) {
		t.Errorf("Invalid. %s", string(body))
		return
	}
}
