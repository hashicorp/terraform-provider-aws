package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/finder"
)

func ProductStatus(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(productID),
		}

		if acceptLanguage != "" {
			input.AcceptLanguage = aws.String(acceptLanguage)
		}

		output, err := conn.DescribeProductAsAdmin(input)

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
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing product status: %w", err)
		}

		if output == nil || output.ProductViewDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("error describing product status: empty product view detail")
		}

		return output, aws.StringValue(output.ProductViewDetail.Status), err
	}
}

func TagOptionStatus(conn *servicecatalog.ServiceCatalog, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(id),
		}

		output, err := conn.DescribeTagOption(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.StatusFailed, fmt.Errorf("error describing tag option: %w", err)
		}

		if output == nil || output.TagOptionDetail == nil {
			return nil, StatusUnavailable, fmt.Errorf("error describing tag option: empty tag option detail")
		}

		return output.TagOptionDetail, servicecatalog.StatusAvailable, err
	}
}

func PortfolioShareStatusWithToken(conn *servicecatalog.ServiceCatalog, token string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.DescribePortfolioShareStatusInput{
			PortfolioShareToken: aws.String(token),
		}
		output, err := conn.DescribePortfolioShareStatus(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.ShareStatusError, fmt.Errorf("error describing portfolio share status: %w", err)
		}

		if output == nil {
			return nil, StatusUnavailable, fmt.Errorf("error describing portfolio share status: empty response")
		}

		return output, aws.StringValue(output.Status), err
	}
}

func PortfolioShareStatus(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.PortfolioShare(conn, portfolioID, shareType, principalID)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, servicecatalog.ShareStatusError, fmt.Errorf("error finding portfolio share: %w", err)
		}

		if output == nil {
			return nil, StatusNotFound, &resource.NotFoundError{
				Message: fmt.Sprintf("error finding portfolio share (%s:%s:%s): empty response", portfolioID, shareType, principalID),
			}
		}

		if !aws.BoolValue(output.Accepted) {
			return output, servicecatalog.ShareStatusInProgress, err
		}

		return output, servicecatalog.ShareStatusCompleted, err
	}
}

func OrganizationsAccessStatus(conn *servicecatalog.ServiceCatalog) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicecatalog.GetAWSOrganizationsAccessStatusInput{}

		output, err := conn.GetAWSOrganizationsAccessStatus(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil, StatusNotFound, err
		}

		if err != nil {
			return nil, OrganizationAccessStatusError, fmt.Errorf("error getting Organizations Access: %w", err)
		}

		if output == nil {
			return nil, StatusUnavailable, fmt.Errorf("error getting Organizations Access: empty response")
		}

		return output, aws.StringValue(output.AccessStatus), err
	}
}
