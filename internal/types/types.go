package types

import "github.com/alfreddobradi/rumour-mill/internal/message"

// Persister is an interface for types that can store data
type Persister interface {
	Persist(message *message.Message) error
}
