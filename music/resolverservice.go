package music

import (
	"log"
	"strings"
)

type Service int

const (
	Spotify Service = iota
	YouTube
)

func ServiceToHumanName(svc Service) string {
	switch svc {
	case Spotify:
		return "Spotify"
	case YouTube:
		return "Youtube"
	}
	return "Unknown"
}

func Service2String(svc Service) string {
	switch svc {
	case Spotify:
		return "sp"
	case YouTube:
		return "yt"
	default:
		return "unk"
	}
}

func String2Service(svc string) Service {
	switch svc {
	case "sp":
		return Spotify
	case "yt":
		return YouTube
	default:
		return -1
	}
}

type ResolverService struct {
	trackResolvers map[Service]TrackResolver
	linkResolvers  map[Service]LinkResolver
}

func NewResolverService() *ResolverService {
	return &ResolverService{
		trackResolvers: make(map[Service]TrackResolver),
		linkResolvers:  make(map[Service]LinkResolver),
	}
}

func (r *ResolverService) RegisterTrackResolver(svc Service, impl TrackResolver) *ResolverService {
	r.trackResolvers[svc] = impl
	return r
}

func (r *ResolverService) RegisterLinkResolver(svc Service, impl LinkResolver) *ResolverService {
	r.linkResolvers[svc] = impl
	return r
}

func (r *ResolverService) ResolveServiceAndId(link string) (string, Service) {
	switch {
	case strings.HasPrefix(link, "https://open.spotify.com/track/"):
		return link[31:], Spotify
	case strings.HasPrefix(link, "https://youtu.be/"):
		return link[17:], YouTube
	}
	return "", -1
}

func (r *ResolverService) ResolveExternalLink(id string) string {
	log.Println("Resolving ", id)
	parts := strings.SplitN(id, ":", 3) // from:to:id
	if len(parts) != 3 {
		return ""
	}
	srcSvc := r.trackResolvers[String2Service(parts[0])]
	if srcSvc == nil {
		return ""
	}
	track, err := srcSvc.ResolveTrack(parts[2])
	if err != nil {
		return ""
	}
	if track == nil {
		return ""
	}
	log.Printf("Resolved track: %#v\n", track)
	dstSvc := r.linkResolvers[String2Service(parts[1])]
	if dstSvc == nil {
		return ""
	}
	link, err := dstSvc.ResolveLink(track)
	log.Println("Resolved link: ", link)
	if err != nil {
		return ""
	}
	return link
}
