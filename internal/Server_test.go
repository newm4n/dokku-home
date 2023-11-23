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

func TestGetActualPath(t *testing.T) {
	assert.Equal(t, "https://www.yahoo.com/", GetActualPath("/api/yahoo", "/api/yahoo", "https://www.yahoo.com", "/"))
	assert.Equal(t, "https://www.yahoo.com/one", GetActualPath("/api/yahoo/one", "/api/yahoo", "https://www.yahoo.com", "/"))
	assert.Equal(t, "https://www.yahoo.com/one/two/three", GetActualPath("/api/yahoo/one/two/three", "/api/yahoo", "https://www.yahoo.com", "/"))

	assert.Equal(t, "https://www.yahoo.com/abc", GetActualPath("/api/yahoo", "/api/yahoo", "https://www.yahoo.com", "/abc"))
	assert.Equal(t, "https://www.yahoo.com/abc/one", GetActualPath("/api/yahoo/one", "/api/yahoo", "https://www.yahoo.com", "/abc"))
	assert.Equal(t, "https://www.yahoo.com/abc/one/two/three", GetActualPath("/api/yahoo/one/two/three", "/api/yahoo", "https://www.yahoo.com", "/abc"))

}
