package codebuild

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	reportGroupStatusUnknown  = "Unknown"
	reportGroupStatusNotFound = "NotFound"
)

// statusReportGroup fetches the Report Group and its Status
func statusReportGroup(conn *codebuild.CodeBuild, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReportGroupByARN(conn, arn)
		if err != nil {
			return nil, reportGroupStatusUnknown, err
		}

		if output == nil {
			return nil, reportGroupStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
