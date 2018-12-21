package play

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrackResolver(t *testing.T) {
	r := NewPlayResolver()
	data, err := os.Open("testdata/sample.html")
	if err != nil {
		t.Fatal(err)
	}
	tracks, err := r.findTrack(data)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "", tracks)
}
