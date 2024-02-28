// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindPortfolioShare(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string) (*types.PortfolioShareDetail, error) {
	input := &servicecatalog.DescribePortfolioSharesInput{
		PortfolioId: aws.String(portfolioID),
		Type:        aws.String(shareType),
	}
	var result *types.PortfolioShareDetail

	err := conn.DescribePortfolioSharesPages(ctx, input, func(page *servicecatalog.DescribePortfolioSharesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.PortfolioShareDetails {
			if deet == nil {
				continue
			}

			if strings.Contains(principalID, aws.ToString(deet.PrincipalId)) {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func FindProductPortfolioAssociation(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string) (*types.PortfolioDetail, error) {
	// seems odd that the sourcePortfolioID is not returned or searchable...
	input := &servicecatalog.ListPortfoliosForProductInput{
		ProductId: aws.String(productID),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	var result *servicecatalog.PortfolioDetail

	err := conn.ListPortfoliosForProductPages(ctx, input, func(page *servicecatalog.ListPortfoliosForProductOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.PortfolioDetails {
			if deet == nil {
				continue
			}

			if aws.ToString(deet.Id) == portfolioID {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func FindBudgetResourceAssociation(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string) (*types.BudgetDetail, error) {
	input := &servicecatalog.ListBudgetsForResourceInput{
		ResourceId: aws.String(resourceID),
	}

	var result *types.BudgetDetail

	err := conn.ListBudgetsForResourcePages(ctx, input, func(page *servicecatalog.ListBudgetsForResourceOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, budget := range page.Budgets {
			if budget == nil {
				continue
			}

			if aws.ToString(budget.BudgetName) == budgetName {
				result = budget
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func FindTagOptionResourceAssociation(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string) (*types.ResourceDetail, error) {
	input := &servicecatalog.ListResourcesForTagOptionInput{
		TagOptionId: aws.String(tagOptionID),
	}

	var result *types.ResourceDetail

	err := conn.ListResourcesForTagOptionPages(ctx, input, func(page *servicecatalog.ListResourcesForTagOptionOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.ResourceDetails {
			if deet == nil {
				continue
			}

			if aws.ToString(deet.Id) == resourceID {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func findProductByID(ctx context.Context, conn *servicecatalog.Client, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	in := &servicecatalog.DescribeProductAsAdminInput{
		Id: aws.String(productID),
	}

	out, err := conn.DescribeProductAsAdmin(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}
