package timescale

import (
	"log"

	"github.com/alfreddobradi/rumour-mill/internal/types"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // driver for postgres
)

// Conn is a connection struct
type Conn struct {
	conn *sqlx.DB
}

// New creates a new TimescaleDB connection and returns it back
func New(uri string) (Conn, error) {
	conn, err := sqlx.Connect("postgres", uri)
	if err != nil {
		return Conn{}, err
	}

	return Conn{conn}, nil
}

// Persist inserts a message to TimescaleDB
func (c *Conn) Persist(message types.Message) (err error) {

	// creates a new transaction value
	tx := c.conn.MustBegin()

	defer func() {
		if err == nil {
			if err = tx.Commit(); err != nil {
				log.Printf("Error while commiting: %v", err)
			}
		} else {
			if err = tx.Rollback(); err != nil {
				log.Printf("Error while commiting: %v", err)
			}
		}
	}()

	query := "INSERT INTO messages VALUES (NOW(), $1, $2)"

	if _, err = tx.Exec(query, message.User, message.Message); err != nil {
		log.Printf("Error while inserting message: %v", err)
	}

	return err
}
