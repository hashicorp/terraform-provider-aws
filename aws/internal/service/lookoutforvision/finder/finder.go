package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lookoutforvision"
)

// ProjectByName returns the project corresponding to the specified name.
// Returns nil if no project is found.
func ProjectByName(conn *lookoutforvision.LookoutForVision, name string) (*lookoutforvision.DescribeProjectOutput, error) {
	input := &lookoutforvision.DescribeProjectInput{
		ProjectName: aws.String(name),
	}

	output, err := conn.DescribeProject(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}