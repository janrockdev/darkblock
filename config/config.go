package config

// Config struct to hold configuration data
type ConfigFile struct {
	NETWORK struct {
		Tick int `mapstructure:"tick"`
	} `mapstructure:"network"`
	KEYS struct {
		GodSeed     string `mapstructure:"god_seed"`
		NodePrivKey string `mapstructure:"node_priv_key"`
	} `mapstructure:"keys"`
}
