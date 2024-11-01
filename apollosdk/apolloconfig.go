package apollosdk

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"
)

type SDKConfig struct {
	AppID            string
	Cluster          string
	MetaAddr         string
	Namespaces       []string
	Secret           string
	IsBackupConfig   bool
	BackupConfigPath string
}

type ApolloSDK struct {
	client agollo.Client
	config SDKConfig
}

func NewApolloSDK(cfg SDKConfig) (*ApolloSDK, error) {
	clientConfig := &config.AppConfig{
		AppID:            cfg.AppID,
		Cluster:          cfg.Cluster,
		IP:               cfg.MetaAddr,
		IsBackupConfig:   cfg.IsBackupConfig,
		BackupConfigPath: cfg.BackupConfigPath,
		Secret:           cfg.Secret,
	}
	clientConfig.NamespaceName = strings.Join(cfg.Namespaces, config.Comma)

	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return clientConfig, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start Apollo client: %v", err)
	}

	sdk := &ApolloSDK{
		client: client,
		config: cfg,
	}

	return sdk, nil
}

func (sdk *ApolloSDK) GetString(namespace, key string) (string, error) {
	config := sdk.client.GetConfig(namespace)
	if config == nil {
		return "", fmt.Errorf("namespace %s not found", namespace)
	}
	return config.GetValue(key), nil
}

func GetConfig[T any](apolloSDK *ApolloSDK, namespace, key string, parseFunc ...func(string) (T, error)) (T, error) {
	var zero T

	value, err := apolloSDK.GetString(namespace, key)
	if err != nil {
		return zero, err
	}

	if len(parseFunc) > 0 && parseFunc[0] != nil {
		return parseFunc[0](value)
	}

	var result T
	if err = json.Unmarshal([]byte(value), &result); err == nil {
		return result, nil
	}

	var str string
	if _, ok := any(result).(string); ok {
		str = value
		return any(str).(T), nil
	}

	return zero, fmt.Errorf("unable to parse config, value: %v", value)
}
