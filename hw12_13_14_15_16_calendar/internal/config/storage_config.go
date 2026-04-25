package config

type StorageConf struct {
	Type string `yaml:"type"` // "memory" or "sql"
	DSN  string `yaml:"dsn"`  // Data Source Name for SQL storage, optional for memory
}
