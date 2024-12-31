package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Setup: create a temporary config file
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "config.yaml")
	configContent := `
app_name: "Darkblock"
version: "1.0.0"
`
	err := os.WriteFile(configFilePath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Set viper to read from the temporary directory
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(tempDir)

	// Test: Load the config
	cfg := LoadConfig()

	// Assertions
	assert.NotNil(t, cfg)
}
