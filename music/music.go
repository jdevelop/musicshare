package music

type Track struct {
	ID     string
	Artist string
	Title  string
	Album  string
}

type Link string

func AsLink(l string) Link {
	return Link(l)
}

type TrackResolver interface {
	ResolveTrack(string) (*Track, error)
}

type LinkResolver interface {
	ResolveLink(*Track) (string, error)
}

type TrackResolverService interface {
	ResolveServiceAndId(string) (string, Service)
	ResolveExternalLink(string) string
}
