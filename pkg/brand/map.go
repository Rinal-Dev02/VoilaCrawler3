package brand

import (
	_ "embed"
	"encoding/json"
	"path/filepath"
)

//go:embed data.json
var brandRawData []byte

var siteBrandMapping = map[string]string{}

func init() {
	if err := json.Unmarshal(brandRawData, &siteBrandMapping); err != nil {
		panic(err)
	}
}

func GetBrand(domain string) string {
	for k, v := range siteBrandMapping {
		if matched, _ := filepath.Match(k, domain); matched {
			return v
		}
	}
	return ""
}
