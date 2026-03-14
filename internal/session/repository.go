package session

// SessionRepository abstracts session persistence so that the service layer
// can be tested without touching the filesystem.
type SessionRepository interface {
	// Load reads the session from storage. A missing file returns an empty
	// session without an error.
	Load() (*Session, error)
	// Save writes session to storage atomically.
	Save(session *Session) error
}
