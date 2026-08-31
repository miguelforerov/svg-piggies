package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type CloudflareAccessVerifier struct {
	client   *http.Client
	verifier *oidc.IDTokenVerifier
}

func NewCloudflareAccessVerifier(
	teamDomain string,
	audience string,
	client *http.Client,
) (*CloudflareAccessVerifier, error) {
	teamDomain = strings.TrimRight(strings.TrimSpace(teamDomain), "/")
	if teamDomain == "" {
		return nil, errors.New("Cloudflare Access team domain is required")
	}
	if strings.TrimSpace(audience) == "" {
		return nil, errors.New("Cloudflare Access audience is required")
	}
	if client == nil {
		return nil, errors.New("HTTP client is required for Cloudflare Access key retrieval")
	}

	ctx := oidc.ClientContext(context.Background(), client)
	keySet := oidc.NewRemoteKeySet(ctx, teamDomain+"/cdn-cgi/access/certs")
	return newCloudflareAccessVerifier(teamDomain, audience, keySet, client), nil
}

func newCloudflareAccessVerifier(
	issuer string,
	audience string,
	keySet oidc.KeySet,
	client *http.Client,
) *CloudflareAccessVerifier {
	return &CloudflareAccessVerifier{
		client: client,
		verifier: oidc.NewVerifier(issuer, keySet, &oidc.Config{
			ClientID:             audience,
			SupportedSigningAlgs: []string{oidc.RS256},
		}),
	}
}

func (v *CloudflareAccessVerifier) Verify(ctx context.Context, rawToken string) error {
	if v.client != nil {
		ctx = oidc.ClientContext(ctx, v.client)
	}
	_, err := v.verifier.Verify(ctx, rawToken)
	return err
}
