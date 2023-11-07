// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
)

func FindConnectionsummaryByName(ctx context.Context, conn *apprunner.Client, name string) (*types.ConnectionSummary, error) {
	input := &apprunner.ListConnectionsInput{
		ConnectionName: aws.String(name),
	}

	var cs *types.ConnectionSummary

	paginator := apprunner.NewListConnectionsPaginator(conn, input, func(o *apprunner.ListConnectionsPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, c := range output.ConnectionSummaryList {
			if c.ConnectionName == nil {
				continue
			}

			if aws.ToString(c.ConnectionName) == name {
				cs = &c
				break
			}
		}
	}

	if cs == nil {
		return nil, nil
	}

	return cs, nil
}

func FindCustomDomain(ctx context.Context, conn *apprunner.Client, domainName, serviceArn string) (*types.CustomDomain, error) {
	input := &apprunner.DescribeCustomDomainsInput{
		ServiceArn: aws.String(serviceArn),
	}

	var customDomain *types.CustomDomain

	paginator := apprunner.NewDescribeCustomDomainsPaginator(conn, input, func(o *apprunner.DescribeCustomDomainsPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, cd := range output.CustomDomains {
			if cd.DomainName == nil {
				continue
			}

			if aws.ToString(cd.DomainName) == domainName {
				customDomain = &cd
				break
			}
		}

	}

	if customDomain == nil {
		return nil, nil
	}

	return customDomain, nil
}
