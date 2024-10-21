package version

import (
	"github.com/txze/goutil/assert"
	"testing"
)

func TestPrint(t *testing.T) {
	err := Print()
	assert.NoError(t, err)
}
