package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config は設定ファイルの構造を表す
type Config struct {
	Level    int            `yaml:"level"`
	Coverage CoverageConfig `yaml:"coverage"`
	Format   string         `yaml:"format"`
	Ignore   []string       `yaml:"ignore"`
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
		Format: "table",
		Ignore: []string{},
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
	if config.Coverage.Min < 0 || config.Coverage.Min > 100 {
		return nil, fmt.Errorf("invalid min coverage: %f (must be between 0 and 100)", config.Coverage.Min)
	}
	if config.Coverage.Max < 0 || config.Coverage.Max > 100 {
		return nil, fmt.Errorf("invalid max coverage: %f (must be between 0 and 100)", config.Coverage.Max)
	}
	if config.Coverage.Min > config.Coverage.Max {
		return nil, fmt.Errorf("min coverage (%f) cannot be greater than max coverage (%f)", config.Coverage.Min, config.Coverage.Max)
	}
	if config.Format != "table" && config.Format != "json" {
		return nil, fmt.Errorf("invalid format: %s (must be 'table' or 'json')", config.Format)
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
func (c *Config) MergeWithFlags(level *int, minCov, maxCov *float64, format *string, ignorePatterns []string) {
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
}
