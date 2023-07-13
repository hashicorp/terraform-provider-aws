// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cur

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
)

func FindReportDefinitionByName(ctx context.Context, conn *costandusagereportservice.CostandUsageReportService, name string) (*costandusagereportservice.ReportDefinition, error) {
	input := &costandusagereportservice.DescribeReportDefinitionsInput{}

	var result *costandusagereportservice.ReportDefinition

	err := conn.DescribeReportDefinitionsPagesWithContext(ctx, input, func(page *costandusagereportservice.DescribeReportDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, reportDefinition := range page.ReportDefinitions {
			if reportDefinition == nil {
				continue
			}

			if aws.StringValue(reportDefinition.ReportName) == name {
				result = reportDefinition
				return false
			}
		}
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
