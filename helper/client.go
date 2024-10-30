package helper

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/linode/linodego"
	"github.com/linode/packer-plugin-linode/version"
	"golang.org/x/oauth2"
)

const TokenEnvVar = "LINODE_TOKEN"

// AddRootCAToTransport applies the CA at the given path to the given *http.Transport
func AddRootCAToTransport(CAPath string, transport *http.Transport) error {
	CAData, err := os.ReadFile(filepath.Clean(CAPath))
	if err != nil {
		return fmt.Errorf("failed to read CA file %s: %w", CAPath, err)
	}

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	if transport.TLSClientConfig.RootCAs == nil {
		transport.TLSClientConfig.RootCAs = x509.NewCertPool()
	}

	transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(CAData)

	return nil
}

func linodeClientFromTransport(transport http.RoundTripper) linodego.Client {
	oauth2Client := &http.Client{
		Transport: transport,
	}

	client := linodego.NewClient(oauth2Client)

	projectURL := "https://www.packer.io"
	userAgent := fmt.Sprintf("Packer/%s (+%s) linodego/%s",
		version.PluginVersion.FormattedVersion(), projectURL, linodego.Version)

	client.SetUserAgent(userAgent)
	return client
}

func getDefaultTransportWithCA(CAPath string) *http.Transport {
	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	AddRootCAToTransport(CAPath, httpTransport)

	return httpTransport
}

func getOauth2TransportWithToken(token string, baseTransport http.RoundTripper) *oauth2.Transport {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauthTransport := &oauth2.Transport{
		Source: tokenSource,
		Base:   baseTransport,
	}
	return oauthTransport
}

func NewLinodeClient(token string) linodego.Client {
	oauthTransport := getOauth2TransportWithToken(token, nil)
	return linodeClientFromTransport(oauthTransport)
}

func NewLinodeClientWithCA(token, CAPath string) linodego.Client {
	oauthTransport := getOauth2TransportWithToken(token, getDefaultTransportWithCA(CAPath))
	return linodeClientFromTransport(oauthTransport)
}
