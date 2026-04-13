package auth

import "encoding/json"

func jsonUnmarshal(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}
