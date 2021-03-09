package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
)

// FileSystemBackupPolicyById returns the EFS Backup Policy corresponding to the specified Id.
// Returns nil if no policy is found.
func FileSystemBackupPolicyById(conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	output, err := conn.DescribeBackupPolicy(&efs.DescribeBackupPolicyInput{
		FileSystemId: aws.String(id),
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.BackupPolicy, nil
}
