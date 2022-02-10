package mq

import (
	"regexp"

	"github.com/beevik/etree"
)

// cannonicalXML reads XML in a string and re-writes it canonically, used for
// comparing XML for logical equivalency
func CanonicalXML(s string) (string, error) {
	doc := etree.NewDocument()
	doc.WriteSettings.CanonicalEndTags = true
	if err := doc.ReadFromString(s); err != nil {
		return "", err
	}

	rawString, err := doc.WriteToString()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`\s`)
	results := re.ReplaceAllString(rawString, "")
	return results, nil
}
