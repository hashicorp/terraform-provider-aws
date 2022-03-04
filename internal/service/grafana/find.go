package grafana

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindLicensedWorkspaceByID(conn *managedgrafana.ManagedGrafana, id string) (*managedgrafana.WorkspaceDescription, error) {
	output, err := FindWorkspaceByID(conn, id)

	if err != nil {
		return nil, err
	}

	if output.LicenseType == nil {
		return nil, &resource.NotFoundError{}
	}

	return output, nil
}

func FindWorkspaceByID(conn *managedgrafana.ManagedGrafana, id string) (*managedgrafana.WorkspaceDescription, error) {
	input := &managedgrafana.DescribeWorkspaceInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspace(input)

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workspace == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workspace, nil
}
