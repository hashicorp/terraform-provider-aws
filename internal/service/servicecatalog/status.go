// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusProduct(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(productID),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeProductAsAdmin(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, err
		}

		if errs.IsA[*awstypes.ResourceInUseException](err) || errs.IsA[*awstypes.LimitExceededException](err) {
			return nil, statusUnavailable, err
		}

		if err != nil {
			return nil, string(awstypes.StatusFailed), fmt.Errorf("describing product status: %w", err)
		}

		if output == nil || output.ProductViewDetail == nil {
			return nil, statusUnavailable, fmt.Errorf("describing product status: empty product view detail")
		}

		return output, string(output.ProductViewDetail.Status), err
	}
}

func statusTagOption(ctx context.Context, conn *servicecatalog.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(id),
		}

		output, err := conn.DescribeTagOption(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, err
		}

		if err != nil {
			return nil, string(awstypes.StatusFailed), fmt.Errorf("describing tag option: %w", err)
		}

		if output == nil || output.TagOptionDetail == nil {
			return nil, statusUnavailable, fmt.Errorf("describing tag option: empty tag option detail")
		}

		return output.TagOptionDetail, string(awstypes.StatusAvailable), err
	}
}

func statusPortfolioShareWithToken(ctx context.Context, conn *servicecatalog.Client, token string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.DescribePortfolioShareStatusInput{
			PortfolioShareToken: aws.String(token),
		}
		output, err := conn.DescribePortfolioShareStatus(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, err
		}

		if err != nil {
			return nil, string(awstypes.ShareStatusError), fmt.Errorf("describing portfolio share status: %w", err)
		}

		if output == nil {
			return nil, statusUnavailable, fmt.Errorf("describing portfolio share status: empty response")
		}

		return output, string(output.Status), err
	}
}

func statusPortfolioShare(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findPortfolioShare(ctx, conn, portfolioID, shareType, principalID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if !output.Accepted {
			return output, string(awstypes.ShareStatusInProgress), nil
		}

		return output, string(awstypes.ShareStatusCompleted), nil
	}
}

func statusOrganizationsAccess(ctx context.Context, conn *servicecatalog.Client) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.GetAWSOrganizationsAccessStatusInput{}

		output, err := conn.GetAWSOrganizationsAccessStatus(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, err
		}

		if err != nil {
			return nil, organizationAccessStatusError, fmt.Errorf("getting Organizations Access: %w", err)
		}

		if output == nil {
			return nil, statusUnavailable, fmt.Errorf("getting Organizations Access: empty response")
		}

		return output, string(output.AccessStatus), err
	}
}

func statusConstraint(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.DescribeConstraintInput{
			Id: aws.String(id),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeConstraint(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("constraint not found (accept language %s, ID: %s): %s", acceptLanguage, id, err),
			}
		}

		if err != nil {
			return nil, string(awstypes.StatusFailed), fmt.Errorf("describing constraint: %w", err)
		}

		if output == nil || output.ConstraintDetail == nil {
			return nil, statusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("describing constraint (accept language %s, ID: %s): empty response", acceptLanguage, id),
			}
		}

		return output, string(output.Status), err
	}
}

func statusProductPortfolioAssociation(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findProductPortfolioAssociation(ctx, conn, acceptLanguage, portfolioID, productID)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("product portfolio association not found (%s): %s", productPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID), err),
			}
		}

		if err != nil {
			return nil, string(awstypes.StatusFailed), fmt.Errorf("describing product portfolio association: %w", err)
		}

		if output == nil {
			return nil, statusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding product portfolio association (%s): empty response", productPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID)),
			}
		}

		return output, string(awstypes.StatusAvailable), err
	}
}

func statusServiceAction(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(id),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeServiceAction(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, err
		}

		if err != nil {
			return nil, string(awstypes.StatusFailed), fmt.Errorf("describing Service Action: %w", err)
		}

		if output == nil || output.ServiceActionDetail == nil {
			return nil, statusUnavailable, fmt.Errorf("describing Service Action: empty Service Action Detail")
		}

		return output.ServiceActionDetail, string(awstypes.StatusAvailable), nil
	}
}

func statusBudgetResourceAssociation(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findBudgetResourceAssociation(ctx, conn, budgetName, resourceID)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("tag option resource association not found (%s): %s", budgetResourceAssociationID(budgetName, resourceID), err),
			}
		}

		if err != nil {
			return nil, string(awstypes.StatusFailed), fmt.Errorf("describing tag option resource association: %w", err)
		}

		if output == nil {
			return nil, statusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding tag option resource association (%s): empty response", budgetResourceAssociationID(budgetName, resourceID)),
			}
		}

		return output, string(awstypes.StatusAvailable), err
	}
}

func statusTagOptionResourceAssociation(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTagOptionResourceAssociation(ctx, conn, tagOptionID, resourceID)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("tag option resource association not found (%s): %s", tagOptionResourceAssociationID(tagOptionID, resourceID), err),
			}
		}

		if err != nil {
			return nil, string(awstypes.StatusFailed), fmt.Errorf("describing tag option resource association: %w", err)
		}

		if output == nil {
			return nil, statusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding tag option resource association (%s): empty response", tagOptionResourceAssociationID(tagOptionID, resourceID)),
			}
		}

		return output, string(awstypes.StatusAvailable), err
	}
}

func statusProvisioningArtifact(ctx context.Context, conn *servicecatalog.Client, id, productID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProvisioningArtifactId: aws.String(id),
			ProductId:              aws.String(productID),
		}

		output, err := conn.DescribeProvisioningArtifact(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, err
		}

		if err != nil {
			return nil, string(awstypes.StatusFailed), err
		}

		if output == nil || output.ProvisioningArtifactDetail == nil {
			return nil, statusUnavailable, err
		}

		return output, string(output.Status), err
	}
}

func statusLaunchPaths(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.ListLaunchPathsInput{
			AcceptLanguage: aws.String(acceptLanguage),
			ProductId:      aws.String(productID),
		}

		var output []awstypes.LaunchPathSummary

		pages := servicecatalog.NewListLaunchPathsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil, statusNotFound, nil
			}

			if err != nil {
				return nil, string(awstypes.StatusFailed), err
			}

			output = append(output, page.LaunchPathSummaries...)
		}

		return output, string(awstypes.StatusAvailable), nil
	}
}

func statusProvisionedProduct(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.DescribeProvisionedProductInput{}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		// one or the other but not both
		if id != "" {
			input.Id = aws.String(id)
		} else if name != "" {
			input.Name = aws.String(name)
		}

		output, err := conn.DescribeProvisionedProduct(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.ProvisionedProductDetail == nil {
			return nil, "", nil
		}

		return output, string(output.ProvisionedProductDetail.Status), err
	}
}

func statusPortfolioConstraints(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &servicecatalog.ListConstraintsForPortfolioInput{
			PortfolioId: aws.String(portfolioID),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		if productID != "" {
			input.ProductId = aws.String(productID)
		}

		var output []awstypes.ConstraintDetail

		pages := servicecatalog.NewListConstraintsForPortfolioPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil, statusNotFound, nil
			}

			if err != nil {
				return nil, string(awstypes.StatusFailed), err
			}

			output = append(output, page.ConstraintDetails...)
		}

		return output, string(awstypes.StatusAvailable), nil
	}
}
