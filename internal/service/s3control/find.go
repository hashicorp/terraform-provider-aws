package s3control

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

func FindMultiRegionAccessPointByAccountIDAndName(conn *s3control.S3Control, accountID string, name string) (*s3control.MultiRegionAccessPointReport, error) {
	input := &s3control.GetMultiRegionAccessPointInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetMultiRegionAccessPoint(input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchMultiRegionAccessPoint) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccessPoint == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccessPoint, nil
}

func findMultiRegionAccessPointOperationByAccountIDAndTokenARN(conn *s3control.S3Control, accountID string, requestTokenARN string) (*s3control.AsyncOperation, error) {
	input := &s3control.DescribeMultiRegionAccessPointOperationInput{
		AccountId:       aws.String(accountID),
		RequestTokenARN: aws.String(requestTokenARN),
	}

	output, err := conn.DescribeMultiRegionAccessPointOperation(input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAsyncRequest) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AsyncOperation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AsyncOperation, nil
}

func FindMultiRegionAccessPointPolicyDocumentByAccountIDAndName(conn *s3control.S3Control, accountID string, name string) (*s3control.MultiRegionAccessPointPolicyDocument, error) {
	input := &s3control.GetMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetMultiRegionAccessPointPolicy(input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchMultiRegionAccessPoint) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policy, nil
}
