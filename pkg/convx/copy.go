package convx

import (
	"encoding/json"
)

func Copy(fromThis, toThis any) error {
	b, err := json.Marshal(fromThis)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &toThis)
}

func Pointer[T any](t T) *T {
	return &t
}
