package spotify

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/jdevelop/musicshare/music"
	"github.com/jdevelop/musicshare/music/spotify/token"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type SpotifyResolver struct {
	client       spotify.Client
	tokenStorage token.TokenStorage
}

var limit = 5

func (s *SpotifyResolver) ResolveLink(track *music.Track) (string, error) {
	var query string
	if track.Title != "" {
		query += track.Title + " "
	}
	res, err := s.client.SearchOpt(query, spotify.SearchTypeTrack, &spotify.Options{
		Limit: &limit,
	})
	if err != nil {
		return "", err
	}
	if len(res.Tracks.Tracks) > 0 {
		return string(res.Tracks.Tracks[0].URI), nil
	}
	return "", nil
}

func (s *SpotifyResolver) ResolveTrack(id string) (*music.Track, error) {
	ft, err := s.client.GetTrack(spotify.ID(id))
	if err != nil {
		return nil, err
	}
	return &music.Track{
		ID:     string(ft.ID),
		Album:  ft.Album.Name,
		Artist: ft.Artists[0].Name, // use the first artist
		Title:  ft.Name,
	}, nil
}

func buildAuth(client, secret, callback string) (*spotify.Authenticator, error) {
	auth := spotify.NewAuthenticator(callback)
	auth.SetAuthInfo(client, secret)
	return &auth, nil
}

func (s *SpotifyResolver) startRefresh() {
	rf := token.NewRefresher(s.tokenStorage, 30*time.Minute, func(*oauth2.Token) (*oauth2.Token, error) {
		return s.client.Token()
	})
	rf.Start()
}

func NewClientToken(client, secret, callback string, ts token.TokenStorage) (*SpotifyResolver, error) {
	auth, err := buildAuth(client, secret, callback)
	if err != nil {
		return nil, err
	}
	t, err := ts.LoadToken()
	if err != nil {
		return nil, err
	}
	c := auth.NewClient(t)
	r := &SpotifyResolver{
		client:       c,
		tokenStorage: ts,
	}
	r.startRefresh()
	return r, nil
}

func NewClient(client, secret, callback string, ts token.TokenStorage) (*SpotifyResolver, error) {
	auth, err := buildAuth(client, secret, callback)
	if err != nil {
		return nil, err
	}
	state := "persistent"
	fmt.Println("Open the following URL:", auth.AuthURL(state))
	fmt.Println("Paste the resulting URL below")
	var (
		sc spotify.Client
	)

	for s := bufio.NewScanner(os.Stdin); s.Scan(); {
		u, err := url.Parse(s.Text())
		if err != nil {
			log.Println("Error parsing URL, please try again", err)
			continue
		}
		if t := u.Query().Get("code"); t == "" {
			log.Printf("No 'code' parameter found in %s\n", s.Text())
			continue
		} else {
			if tkn, err := auth.Exchange(t); err != nil {
				log.Println("Can't exchange token %s => %v\n", t, err)
				continue
			} else {
				if err := ts.SaveToken(tkn); err != nil {
					log.Fatal("Can't save token", err)
				}
				sc = auth.NewClient(tkn)
				break
			}
		}
	}

	r := &SpotifyResolver{
		client:       sc,
		tokenStorage: ts,
	}
	r.startRefresh()
	return r, nil
}

var _ music.TrackResolver = &SpotifyResolver{}
