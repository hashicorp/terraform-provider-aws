package ds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusDirectoryStage(conn *directoryservice.DirectoryService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDirectoryByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Stage), nil
	}
}

func statusDirectoryShare(conn *directoryservice.DirectoryService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDirectoryByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ShareStatus), nil
	}
}

func statusSharedDirectory(ctx context.Context, conn *directoryservice.DirectoryService, ownerId, sharedId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSharedDirectoryByIDs(ctx, conn, ownerId, sharedId)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ShareStatus), nil
	}
}
