package uadmin

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"strings"
)

// {"en": {"key": "en-value"}, "zh": {"key": "zh-value"}}
var languageMap map[string]map[string]string

func getLanguageMap() (m map[string]map[string]string, err error) {
	if len(languageMap) > 0 {
		return languageMap, nil
	}

	fileName := "./templates/uadmin/multilingual.json"
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		Trail(WARNING, "Unrecognized file. fileName:%s", fileName)
	}

	var buf []byte
	if err == nil {
		if buf, err = ioutil.ReadFile(fileName); err != nil {
			Trail(ERROR, "Unable to read language file (%s)", fileName)
		}
	}
	if err == nil {
		if err = json.Unmarshal(buf, &languageMap); err != nil {
			Trail(ERROR, "Unknown translation file format (%s)", fileName)
		}
	}
	return languageMap, err
}

// translation from {multilingual.json}
func getTranslationFromFile(lang, key string) (value string) {
	m, err := getLanguageMap()
	if err == nil {
		if kvs, ok := m[lang]; ok {
			key = strings.Replace(key, " ", "", -1)
			key = strings.ToLower(key)
			value = kvs[key]
		}
	}
	return
}
