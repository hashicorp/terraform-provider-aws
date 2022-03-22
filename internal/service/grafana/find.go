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

func FindSamlConfigurationByID(conn *managedgrafana.ManagedGrafana, id string) (*managedgrafana.SamlAuthentication, error) {
	input := &managedgrafana.DescribeWorkspaceAuthenticationInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspaceAuthentication(input)

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Authentication == nil || output.Authentication.Saml == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := aws.StringValue(output.Authentication.Saml.Status); status == managedgrafana.SamlConfigurationStatusNotConfigured {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output.Authentication.Saml, nil
}

func FindRoleAssociationsByRoleAndWorkspaceID(conn *managedgrafana.ManagedGrafana, role string, workspaceID string) (map[string][]string, error) {
	input := &managedgrafana.ListPermissionsInput{
		WorkspaceId: aws.String(workspaceID),
	}
	output := make(map[string][]string, 0)

	err := conn.ListPermissionsPages(input, func(page *managedgrafana.ListPermissionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Permissions {
			if aws.StringValue(v.Role) == role {
				userType := aws.StringValue(v.User.Type)
				output[userType] = append(output[userType], aws.StringValue(v.User.Id))
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
