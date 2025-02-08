package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/green-ecolution/green-ecolution-backend/pkg/client"
	"golang.org/x/oauth2"
)

var _ oauth2.TokenSource = (*TokenSource)(nil)

type TokenSource struct {
	client *GreenEcolutionClient
	token  *oauth2.Token
}

func NewTokenSource(clientCfg client.Configuration, initToken *oauth2.Token) *TokenSource {
	clientCfg.HTTPClient = http.DefaultClient
	client := NewGreenEcolutionRepo(&clientCfg, slug)

	return &TokenSource{
		client: client,
		token:  initToken,
	}
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	if t.token.Expiry.After(time.Now()) {
		fmt.Println(t.token.RefreshToken)
		newToken, err := t.client.RefreshToken(context.TODO(), t.token.RefreshToken)
		if err != nil {
			return nil, err
		}

		oauthToken, err := t.mapToken(newToken)
		if err != nil {
			return nil, err
		}

		t.token = oauthToken
	}

	return t.token, nil
}

func (t *TokenSource) mapToken(token *client.ClientToken) (*oauth2.Token, error) {
	expiry, err := time.Parse(time.Layout, token.Expiry)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       expiry,
		TokenType:    token.TokenType,
	}, nil
}
