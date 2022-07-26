package ds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findDirectoryByID(conn *directoryservice.DirectoryService, id string) (*directoryservice.DirectoryDescription, error) {
	input := &directoryservice.DescribeDirectoriesInput{
		DirectoryIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeDirectories(input)

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectoryDescriptions) == 0 || output.DirectoryDescriptions[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.DirectoryDescriptions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	directory := output.DirectoryDescriptions[0]

	if stage := aws.StringValue(directory.Stage); stage == directoryservice.DirectoryStageDeleted {
		return nil, &resource.NotFoundError{
			Message:     stage,
			LastRequest: input,
		}
	}

	return directory, nil
}

func findSharedDirectoryByIDs(ctx context.Context, conn *directoryservice.DirectoryService, ownerDirectoryId string, sharedDirectoryId string) (*directoryservice.SharedDirectory, error) { // nosemgrep:ci.ds-in-func-name
	input := &directoryservice.DescribeSharedDirectoriesInput{
		OwnerDirectoryId:   aws.String(ownerDirectoryId),
		SharedDirectoryIds: []*string{aws.String(sharedDirectoryId)},
	}

	output, err := conn.DescribeSharedDirectoriesWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SharedDirectories) == 0 || output.SharedDirectories[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.SharedDirectories); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	sharedDirectory := output.SharedDirectories[0]

	if status := aws.StringValue(sharedDirectory.ShareStatus); status == directoryservice.ShareStatusDeleted {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return sharedDirectory, nil
}
