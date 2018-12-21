package play

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jdevelop/musicshare/music"
)

type playApi struct {
	client *http.Client
}

var (
	cl   http.Client
	once sync.Once
)

func NewPlayResolver() *playApi {
	once.Do(func() {
		cl = http.Client{
			Timeout: 30 * time.Second,
		}
	})
	return &playApi{}
}

func (p *playApi) findTrack(content io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(content)
	if err != nil {
		return "", err
	}
	var (
		link string
	)
	doc.Find("#body-content > div > div > div.main-content > div > div:nth-child(3) > div > div.id-card-list.card-list.two-cards > div:nth-child(1) > div > div.details > a.title").First().Each(func(_ int, s *goquery.Selection) {
		link, _ = s.Attr("href")
	})
	return link, nil
}

func (p *playApi) ResolveLink(trk *music.Track) (string, error) {
	v := url.Values{}
	v.Add("c", "music")
	v.Add("q", fmt.Sprintf(`artist:"%s" song:%s`, trk.Artist, trk.Title))
	resp, err := p.client.Get("https://play.google.com/store/search?" + v.Encode())
	if err != nil {
		return "", err
	}
	if track, err := p.findTrack(resp.Body); err != nil {
		return "", err
	} else {
		return track, nil
	}
}

func (p *playApi) ResolveTrack(lnk string) (*music.Track, error) {
	panic("not implemented")
}

var i playApi
var _ music.LinkResolver = &i
var _ music.TrackResolver = &i
