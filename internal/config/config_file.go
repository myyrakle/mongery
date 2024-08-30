package config

type ConfigFile struct {
	Basedir      string `yaml:"basedir"`
	OutputSuffix string `yaml:"output-suffix"`
}
