package mockx

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yerassyldanay/requestmaker/pkg/convx"
)

func GetUUID(t *testing.T) uuid.UUID {
	generatedUUID, err := uuid.NewUUID()
	require.NoErrorf(t, err, "failed to generate uuid")
	return generatedUUID
}

func Copy(t *testing.T, fromThis, toThis interface{}) {
	require.NoError(t, convx.Copy(fromThis, toThis))
}

func GetBuffer(t *testing.T, val interface{}) io.Reader {
	b, err := json.Marshal(val)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}
