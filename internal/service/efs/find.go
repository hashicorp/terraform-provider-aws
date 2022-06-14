package efs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBackupPolicyByID(conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
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

func FindFileSystemByID(conn *efs.EFS, id string) (*efs.FileSystemDescription, error) {
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

func FindFileSystemPolicyByID(conn *efs.EFS, id string) (*efs.DescribeFileSystemPolicyOutput, error) {
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

func FindReplicationConfigurationByID(conn *efs.EFS, id string) (*efs.ReplicationConfigurationDescription, error) {
	input := &efs.DescribeReplicationConfigurationsInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeReplicationConfigurations(input)

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) || tfawserr.ErrCodeEquals(err, efs.ErrCodeReplicationNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil ||
		len(output.Replications) == 0 ||
		output.Replications[0] == nil ||
		len(output.Replications[0].Destinations) == 0 ||
		output.Replications[0].Destinations[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Replications[0], nil
}
