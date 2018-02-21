package stdout

import (
	"github.com/alfreddobradi/rumour-mill/internal/logger"
	"github.com/alfreddobradi/rumour-mill/internal/message"
)

var log = logger.New()

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
	if len(msg.Recipient) > 0 {
		log.Debugf("server: message: %s %s -> %s : %s", msg.Type, msg.User, msg.Recipient, msg.Message)
	} else {
		log.Debugf("server: message: %s %s -> global : %s", msg.Type, msg.User, msg.Message)
	}

	return err
}
