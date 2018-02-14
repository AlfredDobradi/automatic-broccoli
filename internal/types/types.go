package types

// Message represents a message document
type Message struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

// Persister is an interface for types that can store data
type Persister interface {
	Persist(Message) error
}
