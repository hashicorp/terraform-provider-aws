package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	documentStatusUnknown = "Unknown"
)

func statusAssociation(ctx context.Context, conn *ssm.SSM, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAssociationById(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// Use the Overview.Status field instead of the root-level Status as DescribeAssociation
		// does not appear to return the root-level Status in the API response at this time.
		if output.Overview == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.Overview.Status), nil
	}
}

// statusDocument fetches the Document and its Status
func statusDocument(ctx context.Context, conn *ssm.SSM, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDocumentByName(ctx, conn, name)

		if err != nil {
			return nil, ssm.DocumentStatusFailed, err
		}

		if output == nil {
			return output, documentStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusServiceSetting(ctx context.Context, conn *ssm.SSM, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindServiceSettingByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
