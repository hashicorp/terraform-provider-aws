// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusProduct(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(productID),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeProductAsAdmin(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, err
		}

		if errs.IsA[*types.ResourceInUseException](err) {
			return nil, StatusUnavailable, err
		}

		if errs.IsA[**types.LimitExceededException](err) {
			return nil, StatusUnavailable, err
		}

		if err != nil {
			return nil, string(types.StatusFailed), fmt.Errorf("describing product status: %w", err)
		}

		if output == nil || output.ProductViewDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("describing product status: empty product view detail")
		}

		return output, aws.ToString(output.ProductViewDetail.Status), err
	}
}

func StatusTagOption(ctx context.Context, conn *servicecatalog.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(id),
		}

		output, err := conn.DescribeTagOption(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
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

func StatusPortfolioShareWithToken(ctx context.Context, conn *servicecatalog.Client, token string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribePortfolioShareStatusInput{
			PortfolioShareToken: aws.String(token),
		}
		output, err := conn.DescribePortfolioShareStatus(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, string(types.ShareStatusError), fmt.Errorf("describing portfolio share status: %w", err)
		}

		if output == nil {
			return nil, StatusUnavailable, fmt.Errorf("describing portfolio share status: empty response")
		}

		return output, aws.ToString(output.Status), err
	}
}

func StatusPortfolioShare(ctx context.Context, conn *servicecatalog.Client, portfolioID, shareType, principalID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPortfolioShare(ctx, conn, portfolioID, shareType, principalID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if !aws.ToBool(output.Accepted) {
			return output, servicecatalog.ShareStatusInProgress, nil
		}

		return output, servicecatalog.ShareStatusCompleted, nil
	}
}

func StatusOrganizationsAccess(ctx context.Context, conn *servicecatalog.Client) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.GetAWSOrganizationsAccessStatusInput{}

		output, err := conn.GetAWSOrganizationsAccessStatus(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, OrganizationAccessStatusError, fmt.Errorf("getting Organizations Access: %w", err)
		}

		if output == nil {
			return nil, StatusUnavailable, fmt.Errorf("getting Organizations Access: empty response")
		}

		return output, aws.ToString(output.AccessStatus), err
	}
}

func StatusConstraint(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeConstraintInput{
			Id: aws.String(id),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeConstraint(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("constraint not found (accept language %s, ID: %s): %s", acceptLanguage, id, err),
			}
		}

		if err != nil {
			return nil, string(types.StatusFailed), fmt.Errorf("describing constraint: %w", err)
		}

		if output == nil || output.ConstraintDetail == nil {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("describing constraint (accept language %s, ID: %s): empty response", acceptLanguage, id),
			}
		}

		return output, aws.ToString(output.Status), err
	}
}

func StatusProductPortfolioAssociation(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProductPortfolioAssociation(ctx, conn, acceptLanguage, portfolioID, productID)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("product portfolio association not found (%s): %s", ProductPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID), err),
			}
		}

		if err != nil {
			return nil, string(types.StatusFailed), fmt.Errorf("describing product portfolio association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding product portfolio association (%s): empty response", ProductPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID)),
			}
		}

		return output, string(types.StatusAvailable), err
	}
}

func StatusServiceAction(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(id),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeServiceAction(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, string(types.StatusFailed), fmt.Errorf("describing Service Action: %w", err)
		}

		if output == nil || output.ServiceActionDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("describing Service Action: empty Service Action Detail")
		}

		return output.ServiceActionDetail, string(types.StatusAvailable), nil
	}
}

func StatusBudgetResourceAssociation(ctx context.Context, conn *servicecatalog.Client, budgetName, resourceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindBudgetResourceAssociation(ctx, conn, budgetName, resourceID)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("tag option resource association not found (%s): %s", BudgetResourceAssociationID(budgetName, resourceID), err),
			}
		}

		if err != nil {
			return nil, string(types.StatusFailed), fmt.Errorf("describing tag option resource association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding tag option resource association (%s): empty response", BudgetResourceAssociationID(budgetName, resourceID)),
			}
		}

		return output, string(types.StatusAvailable), err
	}
}

func StatusTagOptionResourceAssociation(ctx context.Context, conn *servicecatalog.Client, tagOptionID, resourceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTagOptionResourceAssociation(ctx, conn, tagOptionID, resourceID)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("tag option resource association not found (%s): %s", TagOptionResourceAssociationID(tagOptionID, resourceID), err),
			}
		}

		if err != nil {
			return nil, string(types.StatusFailed), fmt.Errorf("describing tag option resource association: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &retry.NotFoundError{
				Message: fmt.Sprintf("finding tag option resource association (%s): empty response", TagOptionResourceAssociationID(tagOptionID, resourceID)),
			}
		}

		return output, string(types.StatusAvailable), err
	}
}

func StatusProvisioningArtifact(ctx context.Context, conn *servicecatalog.Client, id, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProvisioningArtifactId: aws.String(id),
			ProductId:              aws.String(productID),
		}

		output, err := conn.DescribeProvisioningArtifact(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, string(types.StatusFailed), err
		}

		if output == nil || output.ProvisioningArtifactDetail == nil {
			return nil, StatusUnavailable, err
		}

		return output, aws.ToString(output.Status), err
	}
}

func StatusLaunchPaths(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, productID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.ListLaunchPathsInput{
			AcceptLanguage: aws.String(acceptLanguage),
			ProductId:      aws.String(productID),
		}

		var summaries []*types.LaunchPathSummary

		err := conn.ListLaunchPathsPages(ctx, input, func(page *servicecatalog.ListLaunchPathsOutput, lastPage bool) bool {
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

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, nil
		}

		if err != nil {
			return nil, string(types.StatusFailed), err
		}

		return summaries, string(types.StatusAvailable), err
	}
}

func StatusProvisionedProduct(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, id, name string) retry.StateRefreshFunc {
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

		output, err := conn.DescribeProvisionedProduct(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.ProvisionedProductDetail == nil {
			return nil, "", nil
		}

		return output, aws.ToString(output.ProvisionedProductDetail.Status), err
	}
}

func StatusPortfolioConstraints(ctx context.Context, conn *servicecatalog.Client, acceptLanguage, portfolioID, productID string) retry.StateRefreshFunc {
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

		var output []*types.ConstraintDetail

		err := conn.ListConstraintsForPortfolioPages(ctx, input, func(page *servicecatalog.ListConstraintsForPortfolioOutput, lastPage bool) bool {
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

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, StatusNotFound, nil
		}

		if err != nil {
			return nil, string(types.StatusFailed), err
		}

		return output, string(types.StatusAvailable), err
	}
}
