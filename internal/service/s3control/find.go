package s3control

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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

func FindAccessPointByAccountIDAndName(conn *s3control.S3Control, accountID string, name string) (*s3control.GetAccessPointOutput, error) {
	input := &s3control.GetAccessPointInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetAccessPoint(input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindAccessPointPolicyAndStatusByAccountIDAndName(conn *s3control.S3Control, accountID string, name string) (string, *s3control.PolicyStatus, error) {
	input1 := &s3control.GetAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output1, err := conn.GetAccessPointPolicy(input1)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input1,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if output1 == nil {
		return "", nil, tfresource.NewEmptyResultError(input1)
	}

	policy := aws.StringValue(output1.Policy)

	if policy == "" {
		return "", nil, tfresource.NewEmptyResultError(input1)
	}

	input2 := &s3control.GetAccessPointPolicyStatusInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output2, err := conn.GetAccessPointPolicyStatus(input2)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input2,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if output2 == nil || output2.PolicyStatus == nil {
		return "", nil, tfresource.NewEmptyResultError(input2)
	}

	return policy, output2.PolicyStatus, nil
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

func FindObjectLambdaAccessPointByAccountIDAndName(conn *s3control.S3Control, accountID string, name string) (*s3control.ObjectLambdaConfiguration, error) {
	input := &s3control.GetAccessPointConfigurationForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetAccessPointConfigurationForObjectLambda(input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Configuration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Configuration, nil
}

func FindObjectLambdaAccessPointPolicyAndStatusByAccountIDAndName(conn *s3control.S3Control, accountID string, name string) (string, *s3control.PolicyStatus, error) {
	input1 := &s3control.GetAccessPointPolicyForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output1, err := conn.GetAccessPointPolicyForObjectLambda(input1)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input1,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if output1 == nil {
		return "", nil, tfresource.NewEmptyResultError(input1)
	}

	policy := aws.StringValue(output1.Policy)

	if policy == "" {
		return "", nil, tfresource.NewEmptyResultError(input1)
	}

	input2 := &s3control.GetAccessPointPolicyStatusForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output2, err := conn.GetAccessPointPolicyStatusForObjectLambda(input2)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input2,
		}
	}

	if err != nil {
		return "", nil, err
	}

	if output2 == nil || output2.PolicyStatus == nil {
		return "", nil, tfresource.NewEmptyResultError(input2)
	}

	return policy, output2.PolicyStatus, nil
}
