package oauth2_client

import "context"

type (
	OAuth2Client interface {
		Callback(ctx context.Context, code string) (result *GoogleUserInfo, err error)
		AuthCodeURL(ctx context.Context, state string) (url string, err error)
	}
)
