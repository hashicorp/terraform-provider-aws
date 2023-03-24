package codebuild

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Deleted
	reportGroupDeleteTimeout = 2 * time.Minute
)

// waitReportGroupDeleted waits for an ReportGroup to return Deleted
func waitReportGroupDeleted(ctx context.Context, conn *codebuild.CodeBuild, arn string) (*codebuild.ReportGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{codebuild.ReportGroupStatusTypeDeleting},
		Target:  []string{},
		Refresh: statusReportGroup(ctx, conn, arn),
		Timeout: reportGroupDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*codebuild.ReportGroup); ok {
		return output, err
	}

	return nil, err
}
