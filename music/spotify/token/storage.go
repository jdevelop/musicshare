package token

import "golang.org/x/oauth2"

type TokenStorage interface {
	LoadToken() (*oauth2.Token, error)
	SaveToken(*oauth2.Token) error
}
