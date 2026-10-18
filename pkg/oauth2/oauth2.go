package oauth2_client

import (
	"context"
	"encoding/json"
	"io"

	"github.com/RandySteven/jlpt-exercise-be/configs"
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
)

func NewOAuth2Client(config *configs.Config) (*oauth2Client, error) {
	googleOauthConfig := &oauth2.Config{
		ClientID:     config.Configs.OAuth2.Google.ClientID,
		ClientSecret: config.Configs.OAuth2.Google.ClientSecret,
		RedirectURL:  config.Configs.OAuth2.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
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
