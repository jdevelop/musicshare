package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/jdevelop/musicshare/music"
	"github.com/jdevelop/musicshare/music/spotify"
	"github.com/jdevelop/musicshare/music/spotify/token"
	"github.com/jdevelop/musicshare/music/youtube"
	"github.com/jdevelop/musicshare/telegram"
)

var (
	tokenPath = flag.String("path", "", "Token storage path")
)

func main() {

	flag.Parse()

	var (
		spf *spotify.SpotifyResolver
		err error
		p   string
	)

	u, err := user.Current()
	switch {
	case *tokenPath != "" && err != nil:
		p = *tokenPath
	case err == nil && *tokenPath == "":
		p = filepath.Join(u.HomeDir, ".musicshare_token")
	case err != nil:
		log.Fatal(err)
	}

	ts, err := token.NewFileStorage(p)
	if err != nil {
		log.Fatal(err)
	}

	token, err := ts.LoadToken()
	if err != nil {
		log.Fatal(err)
	}

	if token == nil {
		log.Println("Generate new token")
		spf, err = spotify.NewClient(os.Getenv("SPF_CLIENT_ID"), os.Getenv("SPF_CLIENT_SECRET"), os.Getenv("SPF_CLIENT_CALLBACK"), ts)
	} else {
		log.Println("Reuse existing token")
		spf, err = spotify.NewClientToken(os.Getenv("SPF_CLIENT_ID"), os.Getenv("SPF_CLIENT_SECRET"), os.Getenv("SPF_CLIENT_CALLBACK"), ts)
	}
	if err != nil {
		log.Fatal(err)
	}

	ytb := youtube.NewYoutubeResolver(os.Getenv("YTB_API_KEY"))

	r := music.NewResolverService().
		RegisterTrackResolver(music.Spotify, spf).
		RegisterLinkResolver(music.Spotify, spf).
		RegisterTrackResolver(music.YouTube, ytb).
		RegisterLinkResolver(music.YouTube, ytb)

	log.Println("Starting TG bot")

	bot := telegram.NewTelegramBot(os.Getenv("TGM_BOT_KEY"))
	if err := bot.Connect(r); err != nil {
		log.Fatal(err)
	}
	log.Println("Started bot, waiting")
	time.Sleep(1 * time.Hour)
}
