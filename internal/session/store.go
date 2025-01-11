package session

type SessionStore interface {
	GetSessions() ([]*Session, error)
	GetSession(sessionId int64) (*Session, error)
	UpdateSession(*Session) error
	CreateSession(*Session) error
}
