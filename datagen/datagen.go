package datagen

import (
	"encoding/json"
	"fmt"

	"github.com/qtumproject/SimpleABI/definitions"
)

func GenerateData(builder definitions.QInterfaceBuilder, input []byte) (string, error) {
	var j interface{}
	json.Unmarshal(input, j)
	m := j.(map[string]interface{})
	fmt.Printf("test: %s\n", m["test"])
	return "", nil
}
