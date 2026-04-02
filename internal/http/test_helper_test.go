package http

import "encoding/json"

func jsonNewDecoder[T any](value string, out *T) error {
	return json.Unmarshal([]byte(value), out)
}
