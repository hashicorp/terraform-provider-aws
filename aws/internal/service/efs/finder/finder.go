package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func BackupPolicyByID(conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	input := &efs.DescribeBackupPolicyInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeBackupPolicy(input)

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BackupPolicy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.BackupPolicy, nil
}

func FileSystemByID(conn *efs.EFS, id string) (*efs.FileSystemDescription, error) {
	input := &efs.DescribeFileSystemsInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeFileSystems(input)

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FileSystems == nil || len(output.FileSystems) == 0 || output.FileSystems[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FileSystems[0], nil
}

func FileSystemPolicyByID(conn *efs.EFS, id string) (*efs.DescribeFileSystemPolicyOutput, error) {
	input := &efs.DescribeFileSystemPolicyInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeFileSystemPolicy(input)

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) || tfawserr.ErrCodeEquals(err, efs.ErrCodePolicyNotFound) {
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
