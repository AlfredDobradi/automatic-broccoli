package stdout

import (
	"log"
	"time"

	"github.com/alfreddobradi/rumour-mill/internal/message"
)

// Conn is a connection struct
type Conn struct {
	conn *bool
}

// New creates a new stdout backend and returns it back
func New(uri string) (Conn, error) {
	conn := true

	return Conn{&conn}, nil
}

// Persist inserts a message to backend
func (c *Conn) Persist(msg *message.Message) (err error) {
	timestampTZ := time.Unix(0, msg.Time).Format(time.RFC3339Nano)

	log.Printf("Message: %s - [%s] %s: %s", timestampTZ, msg.Type, msg.User, msg.Message)

	return err
}
