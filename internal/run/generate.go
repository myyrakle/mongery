package run

import (
	"fmt"

	"github.com/myyrakle/mongery/internal/config"
)

func Generate() {
	configFile := config.Load()

	fmt.Println(configFile.Basedir)
}
