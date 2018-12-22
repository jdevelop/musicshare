package youtube

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jdevelop/musicshare/music"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type YoutubeResolver struct {
	svc *youtube.Service
}

func NewYoutubeResolver(key string) *YoutubeResolver {
	client := &http.Client{
		Transport: &transport.APIKey{Key: key},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}
	return &YoutubeResolver{svc: service}
}

func (y *YoutubeResolver) ResolveLink(t *music.Track) (string, error) {
	var search string
	if t.Title != "" {
		search += "\"" + t.Title + "\" "
	}
	if t.Artist != "" {
		search += "\"" + t.Artist + "\" "
	}
	if t.Album != "" {
		search += "\"" + t.Album + "\" "
	}
	if search == "" {
		return "", nil
	}
	resp, err := y.svc.Search.List("id,snippet").Q(search[:len(search)-1]).Do()
	if err != nil {
		return "", err
	}
	if len(resp.Items) > 0 && resp.Items[0].Id != nil {
		return fmt.Sprintf("https://youtu.be/%s", extractId(resp.Items[0].Id)), nil
	}
	return "", nil
}

func extractId(resId *youtube.ResourceId) string {
	switch {
	case resId.VideoId != "":
		return resId.VideoId
	case resId.PlaylistId != "":
		return resId.PlaylistId
	case resId.ChannelId != "":
		return resId.ChannelId
	}
	return ""
}

var emptyId = errors.New("ID is empty")

func (y *YoutubeResolver) ResolveTrack(id string) (*music.Track, error) {
	if id == "" {
		return nil, emptyId
	}
	resp, err := y.svc.Videos.List("id,snippet").Id(id).Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Items) > 0 {
		r := resp.Items[0]
		return &music.Track{
			Title: r.Snippet.Title,
			ID:    r.Id,
		}, nil
	} else {
		return nil, nil
	}
}

var p = &YoutubeResolver{}

var _ music.TrackResolver = p
var _ music.LinkResolver = p
