package youtube

import (
	"github.com/jdevelop/musicshare/music"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveLink(t *testing.T) {
	s := NewYoutubeResolver(os.Getenv("YTB_API_KEY"))
	lnk, err := s.ResolveLink(&music.Track{
		Artist: "Skrillex",
		Title:  "Bangarang",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Link is ", lnk)
	track, err := s.ResolveTrack("YJVmu6yttiw")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, track)
	assert.Equal(t, "SKRILLEX - Bangarang feat. Sirah [Official Music Video]", track.Title)

}
