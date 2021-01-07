package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
)

func InlinePolicy(ssoadminconn *ssoadmin.SSOAdmin, instanceArn, permissionSetArn string) (*string, error) {
	input := &ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	output, err := ssoadminconn.GetInlinePolicyForPermissionSet(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.InlinePolicy, nil
}
