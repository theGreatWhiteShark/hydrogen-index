package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"
	"github.com/theGreatWhiteShark/hydrogen-index/schema"
)

const schemaResourceName = "hydrogen-index://schema/index.schema.json"

func ValidateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}

	return ValidateBytes(data)
}

func ValidateBytes(data []byte) error {
	compiledSchema, err := compileSchema()
	if err != nil {
		return err
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	if err := compiledSchema.Validate(value); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	var document indexfile.Document
	if err := json.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("decode index.json: %w", err)
	}

	if err := validateCounts(document); err != nil {
		return err
	}

	return nil
}

func compileSchema() (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaResourceName, strings.NewReader(string(schema.IndexJSON))); err != nil {
		return nil, fmt.Errorf("register bundled schema: %w", err)
	}

	compiledSchema, err := compiler.Compile(schemaResourceName)
	if err != nil {
		return nil, fmt.Errorf("compile bundled schema: %w", err)
	}

	return compiledSchema, nil
}

func validateCounts(document indexfile.Document) error {
	if document.PatternCount != len(document.Patterns) {
		return fmt.Errorf("patternCount %d does not match %d pattern entries", document.PatternCount, len(document.Patterns))
	}
	if document.SongCount != len(document.Songs) {
		return fmt.Errorf("songCount %d does not match %d song entries", document.SongCount, len(document.Songs))
	}
	if document.DrumkitCount != len(document.Drumkits) {
		return fmt.Errorf("drumkitCount %d does not match %d drumkit entries", document.DrumkitCount, len(document.Drumkits))
	}
	return nil
}
