package backup

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindVaultAccessPolicyByName(conn *backup.Backup, name string) (*backup.GetBackupVaultAccessPolicyOutput, error) {
	input := &backup.GetBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(name),
	}

	output, err := conn.GetBackupVaultAccessPolicy(input)

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) || tfawserr.ErrCodeEquals(err, errCodeAccessDeniedException) {
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

func FindVaultByName(conn *backup.Backup, name string) (*backup.DescribeBackupVaultOutput, error) {
	input := &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String(name),
	}

	output, err := conn.DescribeBackupVault(input)

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) || tfawserr.ErrCodeEquals(err, errCodeAccessDeniedException) {
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
