package config

type LoggerConf struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"` // file path, empty for stdout
	Format string `yaml:"format"` // "text" or "json"
}
