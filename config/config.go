package config

import "github.com/spf13/viper"

type Config struct {
	SshListenAddr     string
	SshPrivateKeyPath string
	HttpListenAddr    string
	HttpBaseURL       string
}

func NewConfig() *Config {
	viper.AutomaticEnv()

	viper.SetDefault("SSH_LISTEN_ADDR", "0.0.0.0:2222")
	viper.SetDefault("SSH_PRIVATE_KEY_PATH", "./keys/ssh.pem")
	viper.SetDefault("HTTP_LISTEN_ADDR", "0.0.0.0:8080")
	viper.SetDefault("HTTP_BASE_URL", "http://localhost:8080")

	config := &Config{
		SshListenAddr:     viper.GetString("SSH_LISTEN_ADDR"),
		SshPrivateKeyPath: viper.GetString("SSH_PRIVATE_KEY_PATH"),
		HttpListenAddr:    viper.GetString("HTTP_LISTEN_ADDR"),
		HttpBaseURL:       viper.GetString("HTTP_BASE_URL"),
	}

	return config
}
