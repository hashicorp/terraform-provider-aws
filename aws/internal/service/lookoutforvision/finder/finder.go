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

// DatasetByName returns the dataset corresponding to a given project and dataset type
// Returns nil if no dataset is found.
func DatasetByProjectAndType(conn *lookoutforvision.LookoutForVision, project_name string, dataset_type string) (*lookoutforvision.DescribeDatasetOutput, error) {
	input := &lookoutforvision.DescribeDatasetInput{
		ProjectName: aws.String(project_name),
		DatasetType: aws.String(dataset_type),
	}

	output, err := conn.DescribeDataset(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}
