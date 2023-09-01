// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindPortfolioShare(ctx context.Context, conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string) (*servicecatalog.PortfolioShareDetail, error) {
	input := &servicecatalog.DescribePortfolioSharesInput{
		PortfolioId: aws.String(portfolioID),
		Type:        aws.String(shareType),
	}
	var result *servicecatalog.PortfolioShareDetail

	err := conn.DescribePortfolioSharesPagesWithContext(ctx, input, func(page *servicecatalog.DescribePortfolioSharesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.PortfolioShareDetails {
			if deet == nil {
				continue
			}

			if strings.Contains(principalID, aws.StringValue(deet.PrincipalId)) {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return result, nil
}

func FindProductPortfolioAssociation(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) (*servicecatalog.PortfolioDetail, error) {
	// seems odd that the sourcePortfolioID is not returned or searchable...
	input := &servicecatalog.ListPortfoliosForProductInput{
		ProductId: aws.String(productID),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	var result *servicecatalog.PortfolioDetail

	err := conn.ListPortfoliosForProductPagesWithContext(ctx, input, func(page *servicecatalog.ListPortfoliosForProductOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.PortfolioDetails {
			if deet == nil {
				continue
			}

			if aws.StringValue(deet.Id) == portfolioID {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func FindBudgetResourceAssociation(ctx context.Context, conn *servicecatalog.ServiceCatalog, budgetName, resourceID string) (*servicecatalog.BudgetDetail, error) {
	input := &servicecatalog.ListBudgetsForResourceInput{
		ResourceId: aws.String(resourceID),
	}

	var result *servicecatalog.BudgetDetail

	err := conn.ListBudgetsForResourcePagesWithContext(ctx, input, func(page *servicecatalog.ListBudgetsForResourceOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, budget := range page.Budgets {
			if budget == nil {
				continue
			}

			if aws.StringValue(budget.BudgetName) == budgetName {
				result = budget
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func FindTagOptionResourceAssociation(ctx context.Context, conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string) (*servicecatalog.ResourceDetail, error) {
	input := &servicecatalog.ListResourcesForTagOptionInput{
		TagOptionId: aws.String(tagOptionID),
	}

	var result *servicecatalog.ResourceDetail

	err := conn.ListResourcesForTagOptionPagesWithContext(ctx, input, func(page *servicecatalog.ListResourcesForTagOptionOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.ResourceDetails {
			if deet == nil {
				continue
			}

			if aws.StringValue(deet.Id) == resourceID {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func FindPrincipalPortfolioAssociation(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, principalARN, portfolioID string) (*servicecatalog.Principal, error) {
	input := &servicecatalog.ListPrincipalsForPortfolioInput{
		PortfolioId: aws.String(portfolioID),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	var result *servicecatalog.Principal

	err := conn.ListPrincipalsForPortfolioPagesWithContext(ctx, input, func(page *servicecatalog.ListPrincipalsForPortfolioOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.Principals {
			if deet == nil {
				continue
			}

			if aws.StringValue(deet.PrincipalARN) == principalARN {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
