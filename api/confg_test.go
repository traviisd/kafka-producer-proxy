package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAppConfig(t *testing.T) {
	SetAppConfig("../../.helm/files/app-config.json")

	assert.NotNil(t, Config)
}

func TestSetAppConfigError(t *testing.T) {
	err := SetAppConfig("fake")

	assert.NotNil(t, err)
}
