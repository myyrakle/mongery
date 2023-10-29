package annotation

import (
	"strings"
)

type Annotaion struct {
	Name       string
	Parameters []string
}

// @Entity(1, "2", 3) => []string{"1", "2", "3"}
func ParseParameters(commentText string) []string {
	parameters := []string{}

	if !strings.Contains(commentText, "(") {
		return parameters
	}

	parametersText := strings.Split(commentText, "(")[1]

	// find last ')'. if exists, then remove it
	if strings.Contains(parametersText, ")") {
		parametersText = strings.Split(parametersText, ")")[0]
	} else {
		return parameters
	}

	buffer := []byte{}
	for i := 0; i < len(parametersText); i++ {
		if parametersText[i] == ' ' {
			continue
		}

		if parametersText[i] == ',' {
			if len(buffer) > 0 {
				parameters = append(parameters, string(buffer))
				buffer = []byte{}
			}

			continue
		}

		if parametersText[i] == '"' {
			continue
		}

		buffer = append(buffer, parametersText[i])
	}

	if len(buffer) > 0 {
		parameters = append(parameters, string(buffer))
	}

	return parameters
}
