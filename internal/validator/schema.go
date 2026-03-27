// Package validator validates JSON files against the hydrogen-index schema.
package validator

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schema.json
var schemaJSON string

// compiled is the singleton compiled schema, initialised once at package load.
// Panicking here is intentional: a broken embedded schema is a programming
// error, not a runtime condition.
var compiled *jsonschema.Schema

func init() {
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaJSON))
	if err != nil {
		panic(fmt.Sprintf("validator: parse embedded schema: %v", err))
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", doc); err != nil {
		panic(fmt.Sprintf("validator: add embedded schema resource: %v", err))
	}
	compiled, err = c.Compile("schema.json")
	if err != nil {
		panic(fmt.Sprintf("validator: compile embedded schema: %v", err))
	}
}

// Validate validates the JSON file at path against the index schema.
// Returns nil if valid, or a descriptive error listing all failures.
func Validate(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	doc, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		return fmt.Errorf("parse %q: %w", path, err)
	}

	if err := compiled.Validate(doc); err != nil {
		return fmt.Errorf("validation failed:\n%w", err)
	}
	return nil
}
