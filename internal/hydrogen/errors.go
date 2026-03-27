package hydrogen

import "fmt"

func errUnexpectedRoot(path string, actual string, expectedChild string) error {
	return fmt.Errorf("parse %q: root element %q does not contain expected %q data", path, actual, expectedChild)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
