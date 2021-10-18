package s3control

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
)

func findPublicAccessBlockConfiguration(conn *s3control.S3Control, accountID string) (*s3control.PublicAccessBlockConfiguration, error) {
	input := &s3control.GetPublicAccessBlockInput{
		AccountId: aws.String(accountID),
	}

	output, err := conn.GetPublicAccessBlock(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.PublicAccessBlockConfiguration, nil
}

func FindMultiRegionAccessPointByName(conn *s3control.S3Control, accountId string, name string) (*s3control.MultiRegionAccessPointReport, error) {
	input := &s3control.GetMultiRegionAccessPointInput{
		AccountId: aws.String(accountId),
		Name:      aws.String(name),
	}

	log.Printf("[DEBUG] Getting S3 Multi-Region Access Point (%s): %s", name, input)

	output, err := conn.GetMultiRegionAccessPoint(input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccessPoint == nil {
		return nil, nil
	}

	return output.AccessPoint, nil
}

func FindMultiRegionAccessPointPolicyDocumentByName(conn *s3control.S3Control, accountID string, name string) (*s3control.MultiRegionAccessPointPolicyDocument, error) {
	input := &s3control.GetMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetMultiRegionAccessPointPolicy(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.Policy, nil
}
