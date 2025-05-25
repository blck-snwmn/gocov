package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config は設定ファイルの構造を表す
type Config struct {
	Level      int            `yaml:"level"`
	Coverage   CoverageConfig `yaml:"coverage"`
	Format     string         `yaml:"format"`
	Ignore     []string       `yaml:"ignore"`
	Concurrent bool           `yaml:"concurrent"`
	Threshold  float64        `yaml:"threshold"`
}

// CoverageConfig はカバレッジ率フィルタリングの設定
type CoverageConfig struct {
	Min float64 `yaml:"min"`
	Max float64 `yaml:"max"`
}

// DefaultConfig はデフォルトの設定を返す
func DefaultConfig() *Config {
	return &Config{
		Level: 0,
		Coverage: CoverageConfig{
			Min: 0,
			Max: 100,
		},
		Format:     "table",
		Ignore:     []string{},
		Concurrent: false,
		Threshold:  0,
	}
}

// LoadConfig は設定ファイルを読み込む
// ファイルが存在しない場合はnilを返す
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// バリデーション
	if err := ValidateCoverageConfig(config.Coverage.Min, config.Coverage.Max); err != nil {
		return nil, err
	}
	if err := ValidateFormat(config.Format); err != nil {
		return nil, err
	}

	return &config, nil
}

// FindConfigFile は設定ファイルを探す
// カレントディレクトリから親ディレクトリに向かって.gocov.ymlを探す
func FindConfigFile() string {
	configName := ".gocov.yml"

	// カレントディレクトリから開始
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		configPath := filepath.Join(dir, configName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		// 親ディレクトリへ
		parent := filepath.Dir(dir)
		if parent == dir {
			// ルートディレクトリに到達
			break
		}
		dir = parent
	}

	return ""
}

// MergeWithFlags はコマンドライン引数で設定を上書きする
func (c *Config) MergeWithFlags(level *int, minCov, maxCov *float64, format *string, ignorePatterns []string, concurrent *bool, threshold *float64) {
	if level != nil && *level != 0 {
		c.Level = *level
	}
	if minCov != nil && *minCov != 0 {
		c.Coverage.Min = *minCov
	}
	if maxCov != nil && *maxCov != 100 {
		c.Coverage.Max = *maxCov
	}
	if format != nil && *format != "" {
		c.Format = *format
	}
	if len(ignorePatterns) > 0 {
		c.Ignore = ignorePatterns
	}
	if concurrent != nil && *concurrent {
		c.Concurrent = *concurrent
	}
	if threshold != nil && *threshold != 0 {
		c.Threshold = *threshold
	}
}
