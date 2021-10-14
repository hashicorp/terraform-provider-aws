package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// FindReportGroupByARN returns the Report Group corresponding to the specified Arn.
func FindReportGroupByARN(conn *codebuild.CodeBuild, arn string) (*codebuild.ReportGroup, error) {

	output, err := conn.BatchGetReportGroups(&codebuild.BatchGetReportGroupsInput{
		ReportGroupArns: aws.StringSlice([]string{arn}),
	})
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	if len(output.ReportGroups) == 0 {
		return nil, nil
	}

	reportGroup := output.ReportGroups[0]
	if reportGroup == nil {
		return nil, nil
	}

	return reportGroup, nil
}
