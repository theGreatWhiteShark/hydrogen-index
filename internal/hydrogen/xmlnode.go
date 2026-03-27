package hydrogen

import (
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/theGreatWhiteShark/hydrogen-index/internal/indexfile"
)

type ParseOptions struct {
	URL  string
	Size int64
	Hash string
}

type xmlNode struct {
	XMLName  xml.Name
	Content  string    `xml:",chardata"`
	Children []xmlNode `xml:",any"`
}

func parseXMLFile(path string) (xmlNode, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return xmlNode{}, fmt.Errorf("read %q: %w", path, err)
	}

	var root xmlNode
	if err := xml.Unmarshal(data, &root); err != nil {
		return xmlNode{}, fmt.Errorf("parse %q: %w", path, err)
	}

	return root, nil
}

func (node xmlNode) child(localName string) *xmlNode {
	for index := range node.Children {
		if node.Children[index].XMLName.Local == localName {
			return &node.Children[index]
		}
	}
	return nil
}

func (node xmlNode) children(localName string) []xmlNode {
	var matches []xmlNode
	for _, child := range node.Children {
		if child.XMLName.Local == localName {
			matches = append(matches, child)
		}
	}
	return matches
}

func (node xmlNode) text(localName string) string {
	child := node.child(localName)
	if child == nil {
		return ""
	}
	return child.trimmedContent()
}

func (node xmlNode) trimmedContent() string {
	return strings.TrimSpace(node.Content)
}

func textAtPath(node xmlNode, path ...string) string {
	current := &node
	for _, part := range path {
		current = current.child(part)
		if current == nil {
			return ""
		}
	}
	return current.trimmedContent()
}

func nodesAtPath(node xmlNode, path ...string) []xmlNode {
	if len(path) == 0 {
		return []xmlNode{node}
	}

	current := []xmlNode{node}
	for _, part := range path {
		var next []xmlNode
		for _, candidate := range current {
			next = append(next, candidate.children(part)...)
		}
		current = next
		if len(current) == 0 {
			return nil
		}
	}
	return current
}

func collectTexts(node xmlNode, path ...string) []string {
	var values []string
	for _, candidate := range nodesAtPath(node, path...) {
		value := candidate.trimmedContent()
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func parseIntOrDefault(text string, fallback int) int {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return fallback
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return fallback
	}
	return value
}

func uniqueSorted(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func baseArtifact(kind indexfile.ArtifactType, options ParseOptions) indexfile.Artifact {
	return indexfile.Artifact{
		Type: kind,
		URL:  options.URL,
		Hash: options.Hash,
		Size: options.Size,
		Tags: []string{},
	}

}
