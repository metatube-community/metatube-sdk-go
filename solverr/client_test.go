package solverr

import (
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestClient_Get(t *testing.T) {
	c := New("https://flaresolverr-5fjasner4a-an.a.run.app/v1", time.Minute, uuid.New())

	resp, err := c.Get("https://xslist.org/zh")
	if assert.NoError(t, err) {
		data, _ := io.ReadAll(resp.Body)
		t.Logf("Response: %s", data)
	}
}
