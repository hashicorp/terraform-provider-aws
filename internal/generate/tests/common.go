// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tests

import (
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed acctest.tf.gtpl
var acctestTfTmpl string

func AddCommonTemplates(template *template.Template) (*template.Template, error) {
	result, err := template.Parse(acctestTfTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing common \"acctest.tf.gtpl\" config template: %s", err)
	}
	return result, nil
}
