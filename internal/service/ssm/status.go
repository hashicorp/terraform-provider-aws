package ssm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	documentStatusUnknown = "Unknown"
)

func statusAssociation(conn *ssm.SSM, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAssociationById(conn, id)

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
func statusDocument(conn *ssm.SSM, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDocumentByName(conn, name)

		if err != nil {
			return nil, ssm.DocumentStatusFailed, err
		}

		if output == nil {
			return output, documentStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
