// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tests

import (
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed resource_test.go.gtpl
var resourceTestTmpl string

func AddCommonResourceTestTemplates(template *template.Template) (*template.Template, error) {
	result, err := template.Parse(resourceTestTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing common \"resource_test.go.gtpl\" test template: %s", err)
	}
	return result, nil
}

//go:embed acctest.tf.gtpl
var acctestTfTmpl string

func AddCommonTfTemplates(template *template.Template) (*template.Template, error) {
	result, err := template.Parse(acctestTfTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing common \"acctest.tf.gtpl\" config template: %s", err)
	}
	return result, nil
}
