package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppConfig_GetStaticConfig_ReturnsServiceConf(t *testing.T) {
	svc := &Service{}
	cfg := &AppConfig{
		ServiceConf: svc,
	}

	out := cfg.GetStaticConfig()

	assert.NotNil(t, out)
	// GetStaticConfig returns &c.ServiceConf
	assert.Equal(t, &cfg.ServiceConf, out)
}

func TestAppConfig_GetDynamicConfig_ReturnsZkConf(t *testing.T) {
	zk := Zk{}
	cfg := &AppConfig{
		ZkConf: zk,
	}

	out := cfg.GetDynamicConfig()

	assert.NotNil(t, out)
	assert.Same(t, &cfg.ZkConf, out)
}
