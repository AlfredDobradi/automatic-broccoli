package types

// Message represents a message document
type Message struct {
    User    string `json:"user"`
    Message string `json:"message"`
}
