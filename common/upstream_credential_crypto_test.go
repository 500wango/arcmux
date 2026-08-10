package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamCredentialEncryptionRequiresStableSecret(t *testing.T) {
	t.Setenv("UPSTREAM_CREDENTIAL_ENCRYPTION_KEY", "")
	_, err := EncryptUpstreamCredential([]byte("access-token"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UPSTREAM_CREDENTIAL_ENCRYPTION_KEY")
}

func TestUpstreamCredentialEncryptionRoundTripDoesNotExposePlaintext(t *testing.T) {
	t.Setenv("UPSTREAM_CREDENTIAL_ENCRYPTION_KEY", "test-only-stable-secret")
	plaintext := []byte(`{"access_token":"secret-access-token"}`)

	encrypted, err := EncryptUpstreamCredential(plaintext)
	require.NoError(t, err)
	assert.NotContains(t, encrypted, "secret-access-token")

	decrypted, err := DecryptUpstreamCredential(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestUpstreamCredentialEncryptionRejectsWrongSecret(t *testing.T) {
	t.Setenv("UPSTREAM_CREDENTIAL_ENCRYPTION_KEY", "first-secret")
	encrypted, err := EncryptUpstreamCredential([]byte("credential"))
	require.NoError(t, err)

	t.Setenv("UPSTREAM_CREDENTIAL_ENCRYPTION_KEY", "second-secret")
	_, err = DecryptUpstreamCredential(encrypted)
	require.Error(t, err)
}
