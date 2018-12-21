package token

import (
	"log"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type RefreshFunc func(*oauth2.Token) (*oauth2.Token, error)

type Refresher struct {
	tokenStorage TokenStorage
	stop         chan struct{}
	interval     time.Duration
	refreshF     RefreshFunc
	once         sync.Once
}

func NewRefresher(storage TokenStorage, refresh time.Duration, rf RefreshFunc) *Refresher {
	return &Refresher{
		tokenStorage: storage,
		stop:         make(chan struct{}, 1),
		interval:     refresh,
		refreshF:     rf,
	}
}

func (r *Refresher) Close() {
	close(r.stop)
}

func (r *Refresher) Start() {
	r.once.Do(
		func() {
			go func() {
				for {
					select {
					case <-r.stop:
						return
					case <-time.After(r.interval):
						t, err := r.tokenStorage.LoadToken()
						if err != nil {
							log.Println("Can't load token ", err)
							continue
						}
						if t, err := r.refreshF(t); err != nil {
							log.Println("Can't refresh token ", err)
						} else if err := r.tokenStorage.SaveToken(t); err != nil {
							log.Println("Can't save token ", err)
						}
					}
				}
			}()
		})
}
