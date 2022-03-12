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

		return output, aws.StringValue(output.Status.Name), nil
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
