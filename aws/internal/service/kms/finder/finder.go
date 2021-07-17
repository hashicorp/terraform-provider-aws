package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func KeyByID(conn *kms.KMS, id string) (*kms.KeyMetadata, error) {
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(id),
	}

	output, err := conn.DescribeKey(input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KeyMetadata == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	keyMetadata := output.KeyMetadata

	if state := aws.StringValue(keyMetadata.KeyState); state == kms.KeyStatePendingDeletion {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return keyMetadata, nil
}

func KeyPolicyByKeyIDAndPolicyName(conn *kms.KMS, keyID, policyName string) (string, error) {
	input := &kms.GetKeyPolicyInput{
		KeyId:      aws.String(keyID),
		PolicyName: aws.String(policyName),
	}

	output, err := conn.GetKeyPolicy(input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return "", &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Policy == nil {
		return "", &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return aws.StringValue(output.Policy), nil
}
