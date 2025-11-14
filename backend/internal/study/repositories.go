package study

// SessionRepository defines persistence behavior for study sessions.
type SessionRepository interface {
	Create(session StudySession) (StudySession, error)
	Update(session StudySession) (StudySession, error)
	Delete(userID, id string) error
	List(userID string) ([]StudySession, error)
}

// SubjectRepository defines persistence for subjects.
type SubjectRepository interface {
	Create(subject Subject) (Subject, error)
	Update(subject Subject) (Subject, error)
	Delete(userID, id string) error
	List(userID string) ([]Subject, error)
	Get(userID, id string) (Subject, error)
	GetByName(userID, name string) (Subject, error)
}
