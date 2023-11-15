package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExist(t *testing.T) {
	f, err := staticFiles.Open("static/index.html")
	assert.NoError(t, err)
	assert.NotNil(t, f)

	fstat, err := f.Stat()
	assert.NoError(t, err)
	assert.NotNil(t, fstat)
}
