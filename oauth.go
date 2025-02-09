package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/green-ecolution/green-ecolution-backend/pkg/plugin"
	"golang.org/x/oauth2"
)

var _ oauth2.TokenSource = (*TokenSource)(nil)

type TokenSource struct {
	refreshTokenFn func(context.Context, string, string) (*plugin.Token, error)
	token          *oauth2.Token
}

func NewTokenSource(
	refreshTokenFn func(context.Context, string, string) (*plugin.Token, error),
	initToken *oauth2.Token,
) *TokenSource {
	return &TokenSource{
		refreshTokenFn: refreshTokenFn,
		token:          initToken,
	}
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	if t.token.Expiry.Before(time.Now()) {
		slog.Info("refresh token")
		newToken, err := t.refreshTokenFn(context.Background(), cfg.ClientID, cfg.ClientSecret)
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

func (t *TokenSource) mapToken(token *plugin.Token) (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    token.TokenType,
	}, nil
}
