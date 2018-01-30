package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func awsIamUserInlinePolicies(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	results := make([]*schema.ResourceData, 0)
	conn := meta.(*AWSClient).iamconn
	policyNames := make([]*string, 0)
	err := conn.ListUserPoliciesPages(&iam.ListUserPoliciesInput{
		UserName: aws.String(d.Id()),
	}, func(page *iam.ListUserPoliciesOutput, lastPage bool) bool {
		for _, policyName := range page.PolicyNames {
			policyNames = append(policyNames, policyName)
		}

		return lastPage
	})
	if err != nil {
		return nil, err
	}

	for _, policyName := range policyNames {
		policy, err := conn.GetUserPolicy(&iam.GetUserPolicyInput{
			PolicyName: policyName,
			UserName:   aws.String(d.Id()),
		})
		if err != nil {
			return nil, errwrap.Wrapf("Error importing AWS IAM User Policy: {{err}}", err)
		}

		policyResource := resourceAwsIamUserPolicy()
		pData := policyResource.Data(nil)
		pData.SetId(fmt.Sprintf("%s:%s", *policy.UserName, *policy.PolicyName))
		pData.SetType("aws_iam_user_policy")
		pData.Set("name", policy.PolicyName)
		pData.Set("policy", policy.PolicyDocument)
		pData.Set("user", policy.UserName)
		results = append(results, pData)
	}

	return results, nil
}

func resourceAwsIamUserImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	results := make([]*schema.ResourceData, 1)
	results[0] = d

	policyData, err := awsIamUserInlinePolicies(d, meta)
	if err != nil {
		return nil, err
	}

	for _, data := range policyData {
		results = append(results, data)
	}

	return results, nil
}
