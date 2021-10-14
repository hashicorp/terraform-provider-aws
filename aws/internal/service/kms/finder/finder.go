package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func AliasByName(conn *kms.KMS, name string) (*kms.AliasListEntry, error) {
	input := &kms.ListAliasesInput{}
	var output *kms.AliasListEntry

	err := conn.ListAliasesPages(input, func(page *kms.ListAliasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, alias := range page.Aliases {
			if aws.StringValue(alias.AliasName) == name {
				output = alias

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{}
	}

	return output, nil
}

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

	// Once the CMK is in the pending deletion state Terraform considers it logically deleted.
	if state := aws.StringValue(keyMetadata.KeyState); state == kms.KeyStatePendingDeletion {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return keyMetadata, nil
}

func KeyPolicyByKeyIDAndPolicyName(conn *kms.KMS, keyID, policyName string) (*string, error) {
	input := &kms.GetKeyPolicyInput{
		KeyId:      aws.String(keyID),
		PolicyName: aws.String(policyName),
	}

	output, err := conn.GetKeyPolicy(input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Policy, nil
}

func KeyRotationEnabledByKeyID(conn *kms.KMS, keyID string) (*bool, error) {
	input := &kms.GetKeyRotationStatusInput{
		KeyId: aws.String(keyID),
	}

	output, err := conn.GetKeyRotationStatus(input)

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.KeyRotationEnabled, nil
}
