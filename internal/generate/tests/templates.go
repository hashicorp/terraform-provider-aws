// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tests

import (
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed common_test.go.gtpl
var commonTestGoTmpl string

//go:embed resource_test.go.gtpl
var resourceTestGoTmpl string

func AddCommonResourceTestTemplates(template *template.Template) (*template.Template, error) {
	result, err := template.Parse(commonTestGoTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing common \"common_test.go.gtpl\" test template: %w", err)
	}

	result, err = result.Parse(resourceTestGoTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing common \"resource_test.go.gtpl\" test template: %w", err)
	}

	return result, nil
}

//go:embed data_source_test.go.gtpl
var dataSourceTestGoTmpl string

func AddCommonDataSourceTestTemplates(template *template.Template) (*template.Template, error) {
	result, err := template.Parse(commonTestGoTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing common \"common_test.go.gtpl\" test template: %w", err)
	}

	result, err = result.Parse(dataSourceTestGoTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing common \"datasource_test.go.gtpl\" test template: %w", err)
	}

	return result, nil
}

//go:embed acctest.tf.gtpl
var acctestTfTmpl string

func AddCommonTfTemplates(template *template.Template) (*template.Template, error) {
	result, err := template.Parse(acctestTfTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing common \"acctest.tf.gtpl\" config template: %w", err)
	}
	return result, nil
}
