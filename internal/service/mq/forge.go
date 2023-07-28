// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"regexp"

	"github.com/beevik/etree"
)

// CanonicalXML reads XML in a string and re-writes it canonically, used for
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

// CanonicalCuttlefish reads Cuttlefish in a string and re-writes it canonically, used for
// comparing Cuttlefish for logical equivalency
func CanonicalCuttlefish(input string) (string, error) {
	commentRex := regexp.MustCompile(`^\s*(#.*)?$`)
	includeRex := regexp.MustCompile(`^\s*include\s*([A-Za-z0-9-\_\.\*\/]+)\s*(#.*)?`)
	settingRex := regexp.MustCompile(`^\s*([A-Za-z0-9_-]+(\.[A-Za-z0-9_-]+)*)\s*=\s*([^\s]+)\s*(#.*)?`)

	v := []string{}
	s := strings.Split(input, "\n")

	for _, x := range s {
		comment := commentRex.FindStringSubmatch(x)
		include := includeRex.FindStringSubmatch(x)
		setting := settingRex.FindStringSubmatch(x)

		if comment != nil {
			trimmedComment := strings.TrimSpace(comment[1])
			if len(trimmedComment) > 0 {
				v = append(v, trimmedComment)
			}
		}

		if include != nil {
			trimmedComment := strings.TrimSpace(include[2])
			v = append(v, fmt.Sprintf("include %s%s", include[1], trimmedComment))
		}

		if setting != nil {
			trimmedComment := strings.TrimSpace(setting[4])
			v = append(v, fmt.Sprintf("%s=%s%s", setting[1], setting[3], trimmedComment))
		}
	}

	return strings.Join(v[:], "\n"), nil
}
