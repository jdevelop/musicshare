package spotify

import (
	"errors"
	"fmt"
	"github.com/jdevelop/musicshare/music"
	"github.com/jdevelop/musicshare/music/spotify/token"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type hostport struct {
	host string
	port string
}

func buildAuth(client, secret, callback string) (*spotify.Authenticator, *hostport, error) {
	cb, err := url.Parse(callback)
	if err != nil {
		return nil, nil, err
	}
	host := cb.Hostname()
	port := cb.Port()
	if port == "" {
		port = "80"
	}
	auth := spotify.NewAuthenticator(callback)
	auth.SetAuthInfo(client, secret)
	return &auth, &hostport{host: host, port: port}, nil
}

func (s *SpotifyResolver) startRefresh() {
	rf := token.NewRefresher(s.tokenStorage, 30*time.Minute, func(*oauth2.Token) (*oauth2.Token, error) {
		return s.client.Token()
	})
	rf.Start()
}

func NewClientToken(client, secret, callback string, ts token.TokenStorage) (*SpotifyResolver, error) {
	auth, _, err := buildAuth(client, secret, callback)
	if err != nil {
		return nil, err
	}
	t, err := ts.LoadToken()
	if err != nil {
		return nil, err
	}
	c := auth.NewClient(t)
	r := &SpotifyResolver{
		client: c,
	}
	log.Println("1")
	r.startRefresh()
	log.Println("2")
	return r, nil
}

func NewClient(client, secret, callback string, ts token.TokenStorage) (*SpotifyResolver, error) {
	auth, hp, err := buildAuth(client, secret, callback)
	if err != nil {
		return nil, err
	}
	state := RandStringBytes(7)
	fmt.Println(auth.AuthURL(state))
	done := make(chan struct{})

	var (
		sc *spotify.Client
	)

	var serve http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.Token(state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}
		log.Printf("Client token '%s'\n", token.AccessToken)
		if err := ts.SaveToken(token); err != nil {
			log.Println("Can't save token", err)
		}
		// create a client using the specified token
		c := auth.NewClient(token)
		sc = &c
		close(done)
	}

	s := http.Server{
		Addr:    hp.host + ":" + hp.port,
		Handler: serve,
	}
	defer s.Close()

	go func() {
		log.Printf("Starting server at %s:%s\n", hp.host, hp.port)
		s.ListenAndServe()
	}()

	select {
	case <-done:
		r := &SpotifyResolver{
			client: *sc,
		}
		r.startRefresh()
		return r, nil
	case <-time.After(60 * time.Second):
		return nil, errors.New("can't get the response token in 60 seconds")
	}
}

var _ music.TrackResolver = &SpotifyResolver{}
