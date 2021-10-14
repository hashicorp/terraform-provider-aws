package codestarconnections

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
)

// findConnectionByARN returns the Connection corresponding to the specified Arn.
func findConnectionByARN(conn *codestarconnections.CodeStarConnections, arn string) (*codestarconnections.Connection, error) {
	output, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
		ConnectionArn: aws.String(arn),
	})
	if err != nil {
		return nil, err
	}

	if output == nil || output.Connection == nil {
		return nil, nil
	}

	return output.Connection, nil
}
