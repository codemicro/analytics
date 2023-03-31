package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type configLoader struct {
	rawConfigFileContents map[string]any
	lastKey               string
}

func (cl *configLoader) load(fname string) error {
	cl.rawConfigFileContents = make(map[string]any)
	fcont, err := os.ReadFile(fname)
	if err != nil {
		log.Warn().Str("filename", fname).Msg("cannot load config file")
		return nil
	}

	if err := yaml.Unmarshal(fcont, &cl.rawConfigFileContents); err != nil {
		return fmt.Errorf("could not unmarshal config file: %v", err)
	}
	return nil
}

type optionalItem struct {
	item  any
	found bool
}

var indexedPartRegexp = regexp.MustCompile(`(?m)([a-zA-Z]+)(?:\[(\d+)\])?`)

func (cl *configLoader) get(key string) optionalItem {
	// httpcore[2].bananas
	cl.lastKey = key

	parts := strings.Split(key, ".")
	var cursor any = cl.rawConfigFileContents
	for _, part := range parts {
		components := indexedPartRegexp.FindStringSubmatch(part)
		key := components[1]
		index, _ := strconv.ParseInt(components[2], 10, 32)
		isIndexed := components[2] != ""

		item, found := cursor.(map[string]any)[key]
		if !found {
			return optionalItem{nil, false}
		}

		if isIndexed {
			arr, conversionOk := item.([]any)
			if !conversionOk {
				log.Fatal().Msgf("attempted to index non-indexable config item %s", key)
			}
			cursor = arr[index]
		} else {
			cursor = item
		}
	}
	return optionalItem{cursor, true}
}

func (cl *configLoader) required(key string) optionalItem {
	opt := cl.get(key)
	if !opt.found {
		log.Fatal().Msgf("required key %s not found in config file", cl.lastKey)
	}
	return opt
}

func (cl *configLoader) withDefault(key string, defaultValue any) optionalItem {
	opt := cl.get(key)
	if !opt.found {
		return optionalItem{item: defaultValue, found: true}
	}
	return opt
}

func asInt(x optionalItem) int {
	if !x.found {
		return 0
	}
	return x.item.(int)
}

func asString(x optionalItem) string {
	if !x.found {
		return ""
	}
	return x.item.(string)
}

func asBool(x optionalItem) bool {
	if !x.found {
		return false
	}
	return x.item.(bool)
}
