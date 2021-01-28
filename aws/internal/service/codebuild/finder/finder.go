package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
)

// ReportGroupByArn returns the Report Group corresponding to the specified Arn.
func ReportGroupByArn(conn *codebuild.CodeBuild, arn string) (*codebuild.BatchGetReportGroupsOutput, error) {

	output, err := conn.BatchGetReportGroups(&codebuild.BatchGetReportGroupsInput{
		ReportGroupArns: aws.StringSlice([]string{arn}),
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}
