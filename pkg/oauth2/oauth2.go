package oauth2_client

import (
	"context"
	"encoding/json"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type (
	oauth2Client struct {
		googleOauthConfig *oauth2.Config
	}

	GoogleUserInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	OAuth2Config struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
		Scopes       []string
	}
)

func NewOAuth2Client(config *OAuth2Config) (*oauth2Client, error) {
	googleOauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: google.Endpoint,
	}
	return &oauth2Client{googleOauthConfig: googleOauthConfig}, nil
}

func (o *oauth2Client) Callback(ctx context.Context, code string) (result *GoogleUserInfo, err error) {
	token, err := o.googleOauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	client := o.googleOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func (o *oauth2Client) AuthCodeURL(ctx context.Context, state string) (url string, err error) {
	return o.googleOauthConfig.AuthCodeURL(state), nil
}
