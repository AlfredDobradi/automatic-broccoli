package message

import (
	"regexp"
)

// Message represents a message document
type Message struct {
	Type      string `json:"type"`
	Time      int64  `json:"time"`
	User      string `json:"user"`
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

// Parse gets recipient and message from string
func Parse(msg string) map[string]string {
	r, err := regexp.CompilePOSIX(`^@([A-Za-z0-9\-_]+) (.+)`)
	if err != nil {
		return map[string]string{
			"recipient": "",
			"message":   msg,
		}
	}

	matches := r.FindSubmatch([]byte(msg))

	if matches == nil {
		return map[string]string{
			"recipient": "",
			"message":   msg,
		}
	}

	return map[string]string{
		"recipient": string(matches[1]),
		"message":   string(matches[2]),
	}
}
