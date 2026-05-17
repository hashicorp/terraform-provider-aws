// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findPortfolioShare(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string) (*awstypes.PortfolioShareDetail, error) {
	output, err := findPortfolioShares(ctx, conn, portfolioID, shareType, principalID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPortfolioShares(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string) ([]awstypes.PortfolioShareDetail, error) {
	input := &servicecatalog.DescribePortfolioSharesInput{
		PortfolioId: aws.String(portfolioID),
		Type:        awstypes.DescribePortfolioShareType(shareType),
	}
	var result []awstypes.PortfolioShareDetail

	pages := servicecatalog.NewDescribePortfolioSharesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, detail := range page.PortfolioShareDetails {
			if strings.Contains(principalID, aws.ToString(detail.PrincipalId)) {
				result = append(result, detail)
			}
		}
	}

	return result, nil
}

func findProductPortfolioAssociation(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string) (*awstypes.PortfolioDetail, error) {
	output, err := findProductPortfolioAssociations(ctx, conn, acceptLanguage, portfolioID, productID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findProductPortfolioAssociations(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string) ([]awstypes.PortfolioDetail, error) {
	// seems odd that the sourcePortfolioID is not returned or searchable...
	input := &servicecatalog.ListPortfoliosForProductInput{
		ProductId: aws.String(productID),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	var result []awstypes.PortfolioDetail

	pages := servicecatalog.NewListPortfoliosForProductPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, detail := range page.PortfolioDetails {
			if aws.ToString(detail.Id) == portfolioID {
				result = append(result, detail)
			}
		}
	}

	return result, nil
}

func findBudgetResourceAssociation(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string) (*awstypes.BudgetDetail, error) {
	output, err := findBudgetResourceAssociations(ctx, conn, budgetName, resourceID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findBudgetResourceAssociations(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string) ([]awstypes.BudgetDetail, error) {
	input := &servicecatalog.ListBudgetsForResourceInput{
		ResourceId: aws.String(resourceID),
	}

	var result []awstypes.BudgetDetail

	pages := servicecatalog.NewListBudgetsForResourcePaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, budget := range page.Budgets {
			if aws.ToString(budget.BudgetName) == budgetName {
				result = append(result, budget)
			}
		}
	}

	return result, nil
}

func findTagOptionResourceAssociation(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string) (*awstypes.ResourceDetail, error) {
	output, err := findTagOptionResourceAssociations(ctx, conn, tagOptionID, resourceID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTagOptionResourceAssociations(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string) ([]awstypes.ResourceDetail, error) {
	input := &servicecatalog.ListResourcesForTagOptionInput{
		TagOptionId: aws.String(tagOptionID),
	}

	var result []awstypes.ResourceDetail

	pages := servicecatalog.NewListResourcesForTagOptionPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, detail := range page.ResourceDetails {
			if aws.ToString(detail.Id) == resourceID {
				result = append(result, detail)
			}
		}
	}

	return result, nil
}

func findProductByID(ctx context.Context, conn *servicecatalog.Client, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	in := &servicecatalog.DescribeProductAsAdminInput{
		Id: aws.String(productID),
	}

	out, err := conn.DescribeProductAsAdmin(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}
