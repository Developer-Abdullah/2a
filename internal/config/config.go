package config

import (
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	Database struct {
		DSN string `mapstructure:"DB_DSN"`
	}
	Redis struct {
		URL string `mapstructure:"REDIS_URL"`
	}
	S3 struct {
		Endpoint, AccessKey, SecretKey, Region, Bucket string `mapstructure:"S3_ENDPOINT"`
	}
	Security struct{ AESEncryptionKey, JWTPrivateKeyPEM, JWTPublicKeyPEM string }
	APNs     struct{ KeyID, TeamID, PrivateKeyPEM string }
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("S3_REGION", "us-east-1")
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
