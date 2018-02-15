package avro

import (
	"encoding/json"

	"github.com/alfreddobradi/rumour-mill/internal/types"
	goavro "github.com/linkedin/goavro"
)

var schema = `
    {
        "type": "record",
        "name": "Message",
        "fields": [
            { "name": "time", "type": "long" },
            { "name": "user", "type": "string" },
            { "name": "message", "type": "string" }
        ]
    }
`

// Encode takes a message and encodes it to avro format
func Encode(msg types.Message) ([]uint8, error) {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, err
	}

	textual, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	native, _, err := codec.NativeFromTextual(textual)
	if err != nil {
		return nil, err
	}

	binary, err := codec.BinaryFromNative(nil, native)
	return binary, err

}

// Decode takes a binary avro buffer and decodes it into a message
func Decode(buf []byte) (m types.Message, err error) {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return
	}

	native, _, err := codec.NativeFromBinary(buf)
	if err != nil {
		return
	}

	textual, err := codec.TextualFromNative(nil, native)
	if err != nil {
		return
	}

	err = json.Unmarshal(textual, &m)
	if err != nil {
		return
	}

	return
}
