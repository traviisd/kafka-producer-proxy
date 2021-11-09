package api

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAppSecrets(t *testing.T) {
	b, err := ioutil.ReadFile("../testdata/secrets.json")
	if err != nil {
		t.Fail()
	}
	SetAppSecrets(b)

	assert.NotNil(t, Secrets)
	assert.True(t, len(Secrets.OAuthClientSecret) > 0)
}
