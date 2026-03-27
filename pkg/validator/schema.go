package validator

import (
	_ "embed"
)

// IndexSchema contains the JSON schema for validating the index file.
//go:embed schema.json
var IndexSchema string
