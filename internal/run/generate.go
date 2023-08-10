package run

import (
	"github.com/myyrakle/mongery/internal/config"
	"github.com/myyrakle/mongery/pkg"
)

func RunGenerate() {
	configFile := config.Load()

	pkg.Generate(configFile)
}
