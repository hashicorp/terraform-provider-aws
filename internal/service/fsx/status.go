package fsx

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func statusAdministrativeAction(conn *fsx.FSx, fsID, actionType string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAdministrativeActionByFileSystemIDAndActionType(conn, fsID, actionType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusBackup(conn *fsx.FSx, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindBackupByID(conn, id)

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
		output, err := FindFileSystemByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Lifecycle), nil
	}
}
