package config

type Feature string

type Features []Feature

func (f Features) Contains(feature Feature) bool {
	for _, v := range f {
		if v == feature {
			return true
		}
	}
	return false
}

const (
	FeatureSlice Feature = "SLICE"
)

type ConfigFile struct {
	Basedir      string   `yaml:"basedir"`
	OutputSuffix string   `yaml:"output-suffix"`
	Features     Features `yaml:"features"` // 사용할 기능 목록
}
