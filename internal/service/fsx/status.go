package fsx

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/fsx/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func statusBackup(conn *fsx.FSx, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindBackupByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Lifecycle), nil
	}
}

func statusFileSystem(conn *fsx.FSx, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindFileSystemByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Lifecycle), nil
	}
}

func statusFileSystemAdministrativeActions(conn *fsx.FSx, id, action string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindFileSystemByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, administrativeAction := range output.AdministrativeActions {
			if administrativeAction == nil {
				continue
			}

			if aws.StringValue(administrativeAction.AdministrativeActionType) == action {
				return output, aws.StringValue(administrativeAction.Status), nil
			}
		}

		return output, fsx.StatusCompleted, nil
	}
}
