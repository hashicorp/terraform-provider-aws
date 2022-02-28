package workmail

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindOrganizationByID(conn *workmail.WorkMail, id string) (*workmail.DescribeOrganizationOutput, error) {
	input := &workmail.DescribeOrganizationInput{
		OrganizationId: aws.String(id),
	}

	output, err := conn.DescribeOrganization(input)

	if tfawserr.ErrCodeEquals(err, workmail.ErrCodeOrganizationNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
