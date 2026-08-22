package aws

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"

	"transcode-demo/pkg/utils"
)

type Config struct {
	AccessKeyID      string `mapstructure:"access_key_id"`
	SecretAccessKey  string `mapstructure:"secret_access_key"`
	SessionToken     string `mapstructure:"session_token"`
	Region           string `mapstructure:"region"`
	S3ForcePathStyle bool   `mapstructure:"s3_force_path_style"`
	Endpoint         string `mapstructure:"endpoint"`
}

func (c *Config) SetDefault(parentKey string) {
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "access_key_id"), "test")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "secret_access_key"), "test")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "session_token"), "")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "region"), "us-east-1")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "s3_force_path_style"), true)
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "endpoint"), "http://localhost:4566")
}

var (
	once   sync.Once
	client *s3.Client
)

func (c *Config) GetClient(ctx context.Context) (*s3.Client, error) {
	var err error
	var cfg aws.Config
	once.Do(func() {
		cfg, err = config.LoadDefaultConfig(
			ctx,
			config.WithRegion(c.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, c.SessionToken)),
		)
		cfg.BaseEndpoint = &c.Endpoint
		if err != nil {
			panic(err)
		}
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = c.S3ForcePathStyle
		})
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
