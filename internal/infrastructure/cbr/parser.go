package cbr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

func extractKeyRate(xmlBody string) (string, error) {
	document := etree.NewDocument()

	if err := document.ReadFromString(xmlBody); err != nil {
		return "", fmt.Errorf("parse cbr xml: %w", err)
	}

	candidates := []string{}

	var walk func(element *etree.Element)
	walk = func(element *etree.Element) {
		tag := strings.ToLower(element.Tag)
		text := strings.TrimSpace(element.Text())

		if text != "" && (tag == "rate" || tag == "keyrate" || tag == "kr") {
			candidates = append(candidates, text)
		}

		for _, child := range element.ChildElements() {
			walk(child)
		}
	}

	if root := document.Root(); root != nil {
		walk(root)
	}

	for i := len(candidates) - 1; i >= 0; i-- {
		normalized := strings.ReplaceAll(candidates[i], ",", ".")

		value, err := strconv.ParseFloat(normalized, 64)
		if err == nil && value >= 0 {
			return fmt.Sprintf("%.4f", value), nil
		}
	}

	return "", fmt.Errorf("key rate not found in cbr response")
}
