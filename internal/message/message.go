package message

// Message represents a message document
type Message struct {
	Type    string `json:"type"`
	Time    int64  `json:"time"`
	User    string `json:"user"`
	Message string `json:"message"`
}
