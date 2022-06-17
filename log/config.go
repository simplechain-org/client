package log

type Config struct {
	WriteFile    bool   `yaml:"writeFile"`
	FileRoot     string `yaml:"fileRoot"`
	FilePath     string `yaml:"filePath"`
	MaxBytesSize uint   `yaml:"maxBytesSize"`
	Level        string `yaml:"level"`
}
