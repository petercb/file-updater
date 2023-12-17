// Adapted from https://github.com/open-policy-agent/conftest/tree/v0.47.0/internal/registry
package registry

import (
	"context"

	"github.com/cpuguy83/dockercfg"
	"github.com/petercb/file-updater/internal/network"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func SetupClient(repository *remote.Repository) {
	registry := repository.Reference.Host()

	if network.IsLoopback(network.Hostname(registry)) {
		// Docker by default accesses localhost using plaintext HTTP
		repository.PlainHTTP = true
	}

	client := auth.DefaultClient
	client.SetUserAgent("file-updater")
	client.Credential = func(ctx context.Context, registry string) (auth.Credential, error) {
		host := dockercfg.ResolveRegistryHost(registry)
		username, password, err := dockercfg.GetRegistryCredentials(host)
		if err != nil {
			return auth.EmptyCredential, err
		}

		return auth.Credential{
			Username: username,
			Password: password,
		}, nil
	}

	repository.Client = client
}
