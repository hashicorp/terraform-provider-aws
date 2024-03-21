// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusProduct(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(productID),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeProductAsAdminWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceInUseException) {
			return nil, StatusUnavailable, err
		}

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeLimitExceededException) {
			return nil, StatusUnavailable, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("describing product status: %w", err)
		}

		if output == nil || output.ProductViewDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("describing product status: empty product view detail")
		}

		return output, aws.StringValue(output.ProductViewDetail.Status), err
	}
}

func StatusTagOption(ctx context.Context, conn *servicecatalog.ServiceCatalog, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(id),
		}

		output, err := conn.DescribeTagOptionWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("describing tag option: %w", err)
		}

		if output == nil || output.TagOptionDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("describing tag option: empty tag option detail")
		}

		return output.TagOptionDetail, servicecatalog.StatusAvailable, err
	}
}

func StatusPortfolioShareWithToken(ctx context.Context, conn *servicecatalog.ServiceCatalog, token string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribePortfolioShareStatusInput{
			PortfolioShareToken: aws.String(token),
		}
		output, err := conn.DescribePortfolioShareStatusWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.ShareStatusError, fmt.Errorf("describing portfolio share status: %w", err)
		}

		if output == nil {
			return nil, StatusUnavailable, fmt.Errorf("describing portfolio share status: empty response")
		}

		status := aws.StringValue(output.Status)
		if (status == servicecatalog.ShareStatusCompletedWithErrors || status == servicecatalog.ShareStatusError) && output.ShareDetails != nil && output.ShareDetails.ShareErrors != nil && len(output.ShareDetails.ShareErrors) > 0 {
			return output, status, fmt.Errorf("error with portfolio share status: %+v", output.ShareDetails.ShareErrors)
		}

		return output, status, err
	}
}

func StatusPortfolioShare(ctx context.Context, conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPortfolioShare(ctx, conn, portfolioID, shareType, principalID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if !aws.BoolValue(output.Accepted) {
			return output, servicecatalog.ShareStatusInProgress, nil
		}

		return output, servicecatalog.ShareStatusCompleted, nil
	}
}

func StatusOrganizationsAccess(ctx context.Context, conn *servicecatalog.ServiceCatalog) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.GetAWSOrganizationsAccessStatusInput{}

		output, err := conn.GetAWSOrganizationsAccessStatusWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, OrganizationAccessStatusError, fmt.Errorf("getting Organizations Access: %w", err)
		}

		if output == nil {
			return nil, StatusUnavailable, fmt.Errorf("getting Organizations Access: empty response")
		}

		return output, aws.StringValue(output.AccessStatus), err
	}
}

func StatusConstraint(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeConstraintInput{
			Id: aws.String(id),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeConstraintWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("constraint not found (accept language %s, ID: %s): %s", acceptLanguage, id, err),
			}
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("describing constraint: %w", err)
		}

		if output == nil || output.ConstraintDetail == nil {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("describing constraint (accept language %s, ID: %s): empty response", acceptLanguage, id),
			}
		}

		return output, aws.StringValue(output.Status), err
	}
}

func StatusProductPortfolioAssociation(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProductPortfolioAssociation(ctx, conn, acceptLanguage, portfolioID, productID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("product portfolio association not found (%s): %s", ProductPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID), err),
			}
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("describing product portfolio association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding product portfolio association (%s): empty response", ProductPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID)),
			}
		}

		return output, servicecatalog.StatusAvailable, err
	}
}

func StatusServiceAction(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(id),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeServiceActionWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("describing Service Action: %w", err)
		}

		if output == nil || output.ServiceActionDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("describing Service Action: empty Service Action Detail")
		}

		return output.ServiceActionDetail, servicecatalog.StatusAvailable, nil
	}
}

func StatusBudgetResourceAssociation(ctx context.Context, conn *servicecatalog.ServiceCatalog, budgetName, resourceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindBudgetResourceAssociation(ctx, conn, budgetName, resourceID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("tag option resource association not found (%s): %s", BudgetResourceAssociationID(budgetName, resourceID), err),
			}
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("describing tag option resource association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding tag option resource association (%s): empty response", BudgetResourceAssociationID(budgetName, resourceID)),
			}
		}

		return output, servicecatalog.StatusAvailable, err
	}
}

func StatusTagOptionResourceAssociation(ctx context.Context, conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTagOptionResourceAssociation(ctx, conn, tagOptionID, resourceID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("tag option resource association not found (%s): %s", TagOptionResourceAssociationID(tagOptionID, resourceID), err),
			}
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("describing tag option resource association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding tag option resource association (%s): empty response", TagOptionResourceAssociationID(tagOptionID, resourceID)),
			}
		}

		return output, servicecatalog.StatusAvailable, err
	}
}

func StatusProvisioningArtifact(ctx context.Context, conn *servicecatalog.ServiceCatalog, id, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProvisioningArtifactId: aws.String(id),
			ProductId:              aws.String(productID),
		}

		output, err := conn.DescribeProvisioningArtifactWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, err
		}

		if output == nil || output.ProvisioningArtifactDetail == nil {
			return nil, StatusUnavailable, err
		}

		return output, aws.StringValue(output.Status), err
	}
}

func StatusLaunchPaths(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.ListLaunchPathsInput{
			AcceptLanguage: aws.String(acceptLanguage),
			ProductId:      aws.String(productID),
		}

		var summaries []*servicecatalog.LaunchPathSummary

		err := conn.ListLaunchPathsPagesWithContext(ctx, input, func(page *servicecatalog.ListLaunchPathsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, summary := range page.LaunchPathSummaries {
				if summary == nil {
					continue
				}

				summaries = append(summaries, summary)
			}

			return !lastPage
		})

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, nil
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, err
		}

		return summaries, servicecatalog.StatusAvailable, err
	}
}

func StatusProvisionedProduct(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, id, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
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

		output, err := conn.DescribeProvisionedProductWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.ProvisionedProductDetail == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.ProvisionedProductDetail.Status), err
	}
}

func StatusPortfolioConstraints(ctx context.Context, conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.ListConstraintsForPortfolioInput{
			PortfolioId: aws.String(portfolioID),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		if productID != "" {
			input.ProductId = aws.String(productID)
		}

		var output []*servicecatalog.ConstraintDetail

		err := conn.ListConstraintsForPortfolioPagesWithContext(ctx, input, func(page *servicecatalog.ListConstraintsForPortfolioOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, deet := range page.ConstraintDetails {
				if deet == nil {
					continue
				}

				output = append(output, deet)
			}

			return !lastPage
		})

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, nil
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, err
		}

		return output, servicecatalog.StatusAvailable, err
	}
}
