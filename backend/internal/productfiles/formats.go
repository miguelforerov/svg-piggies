package productfiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

//go:generate node ../../../scripts/generate-product-file-formats.mjs

type formatDefinition struct {
	Format       string   `json:"format"`
	Extension    string   `json:"extension"`
	ContentTypes []string `json:"contentTypes"`
}

type formatConfiguration struct {
	Comment json.RawMessage    `json:"_comment"`
	Formats []formatDefinition `json:"formats"`
}

var formatIDPattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
var contentTypePattern = regexp.MustCompile(`^[a-z0-9.+-]+/[a-z0-9.+-]+$`)

func parseFormatConfiguration(data []byte) (map[string]formatDefinition, error) {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var config formatConfiguration
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("decode product file formats: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errors.New("product file formats contain trailing data")
	}
	if len(config.Comment) > 0 {
		var comment string
		if string(config.Comment) == "null" || json.Unmarshal(config.Comment, &comment) != nil {
			return nil, errors.New("product file format comment must be a string")
		}
	}
	if len(config.Formats) == 0 {
		return nil, errors.New("product file formats are empty")
	}
	formats := make(map[string]formatDefinition, len(config.Formats))
	for _, definition := range config.Formats {
		if !formatIDPattern.MatchString(definition.Format) || definition.Extension != "."+definition.Format ||
			len(definition.ContentTypes) == 0 {
			return nil, errors.New("invalid product file format definition")
		}
		if _, exists := formats[definition.Format]; exists {
			return nil, errors.New("duplicate product file format")
		}
		seen := make(map[string]bool, len(definition.ContentTypes))
		for _, contentType := range definition.ContentTypes {
			if !contentTypePattern.MatchString(contentType) || seen[contentType] {
				return nil, errors.New("invalid or duplicate product file content type")
			}
			seen[contentType] = true
		}
		formats[definition.Format] = definition
	}
	return formats, nil
}

var formats, formatConfigurationError = parseFormatConfiguration([]byte(embeddedProductFileFormatsJSON))
