package codebuild

import (
	"time"

	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for an Operation to return Deleted
	reportGroupDeleteTimeout = 2 * time.Minute
)

// waitReportGroupDeleted waits for an ReportGroup to return Deleted
func waitReportGroupDeleted(conn *codebuild.CodeBuild, arn string) (*codebuild.ReportGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{codebuild.ReportGroupStatusTypeDeleting},
		Target:  []string{},
		Refresh: statusReportGroup(conn, arn),
		Timeout: reportGroupDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*codebuild.ReportGroup); ok {
		return output, err
	}

	return nil, err
}
