package registry

import (
	"context"
	"fmt"
	"testing"

	"github.com/petercb/file-updater/internal/network"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func mustParseReference(ref string) *remote.Repository {
	r, err := remote.NewRepository(ref)

	if err == nil {
		return r
	}

	panic(fmt.Sprintf("Unable to parse reference: %s: %v", ref, err))
}

func TestSetupClient(t *testing.T) {
	cases := []struct {
		ref         string
		plaintext   bool
		credentials auth.Credential
	}{
		{ref: "localhost/test:tag", plaintext: true},
		{ref: "127.0.0.1/test:tag", plaintext: true},
		{ref: "127.0.0.1.nip.io/test:tag", plaintext: true},
		{ref: "test.127.0.0.1.nip.io/test:tag", plaintext: true},
		{ref: "test.registry.io/test:tag", plaintext: false, credentials: auth.Credential{Username: "test", Password: "supersecret"}},
	}

	for _, c := range cases {
		t.Run(c.ref, func(t *testing.T) {
			repository := mustParseReference(c.ref)

			t.Setenv("DOCKER_CONFIG", "testdata")

			SetupClient(repository)

			if repository.PlainHTTP != c.plaintext {
				t.Errorf(`expecting repository.PlainHTTP == %v, but it was %v`, c.plaintext, repository.PlainHTTP)
			}

			client, ok := repository.Client.(*auth.Client)
			if !ok {
				t.Error("expecting repository.Client to be instance of `auth.Client`")
			}

			if got := client.Header.Get("User-Agent"); got != "file-updater" {
				t.Errorf(`expecting client.Header['User-Agent'] == "file-updater", but it was %v`, got)
			}

			if got, err := client.Credential(context.Background(), network.Hostname(c.ref)); err == nil {
				if got != c.credentials {
					t.Errorf(`expecting client.Credential(_, "%v") == "%v", but it was %v`, c.ref, c.credentials, got)
				}
			} else {
				t.Errorf(`unexpected error while fetching credentials: %v`, err)
			}
		})
	}
}
