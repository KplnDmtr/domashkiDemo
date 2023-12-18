package handlerstestsutils

import (
	"encoding/json"
	"testing"
)

func ConvertToJSON(t *testing.T, data interface{}) []byte {
	bytes, err := json.Marshal(data)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
		return nil
	}
	return bytes
}
