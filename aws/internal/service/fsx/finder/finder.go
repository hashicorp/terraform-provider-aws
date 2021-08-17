package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func BackupByID(conn *fsx.FSx, id string) (*fsx.Backup, error) {
	input := &fsx.DescribeBackupsInput{
		BackupIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeBackups(input)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) || tfawserr.ErrCodeEquals(err, fsx.ErrCodeBackupNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Backups) == 0 || output.Backups[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Backups[0], nil
}
