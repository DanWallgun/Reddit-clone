package session

import (
	"database/sql"
	"time"
)

type SessionsManager struct {
	DB *sql.DB
}

func NewSessionsManager(db *sql.DB) *SessionsManager {
	return &SessionsManager{
		DB: db,
	}
}

func (sm *SessionsManager) Check(sessionID string) (*Session, error) {
	sess := &Session{}
	var expires_unix int64
	err := sm.DB.
		QueryRow("SELECT id, user_id, expires FROM sessions WHERE id = ?", sessionID).
		Scan(&sess.ID, &sess.UserID, &expires_unix)
	if err != nil {
		return nil, ErrNoAuth
	}
	sess.Expires = time.Unix(expires_unix, 0)
	if sess.Expires.Unix() < time.Now().Unix() {
		return nil, ErrNoAuth
	}

	return sess, nil
}

func (sm *SessionsManager) Create(userID string) (*Session, error) {
	sess, err := NewSession(userID)
	if err != nil {
		return nil, err
	}
	_, err = sm.DB.Exec(
		"INSERT INTO sessions (`id`, `user_id`, `expires`) VALUES (?, ?, ?)",
		sess.ID,
		sess.UserID,
		sess.Expires.Unix(),
	)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

//func (sm *SessionsManager) DestroyCurrent(w http.ResponseWriter, r *http.Request) error {
//	sess, err := SessionFromContext(r.Context())
//	if err != nil {
//		return err
//	}
//
//	sm.mu.Lock()
//	delete(sm.data, sess.ID)
//	sm.mu.Unlock()
//
//	cookie := http.Cookie{
//		Name:    "session_id",
//		Expires: time.Now().AddDate(0, 0, -1),
//		Path:    "/",
//	}
//	http.SetCookie(w, &cookie)
//	return nil
//}
