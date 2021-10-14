package organizations

import (
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func FindOrganization(conn *organizations.Organizations) (*organizations.Organization, error) {
	input := &organizations.DescribeOrganizationInput{}

	output, err := conn.DescribeOrganization(input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Organization == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Organization, nil
}
