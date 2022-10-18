package organizations

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAccountByID(conn *organizations.Organizations, id string) (*organizations.Account, error) {
	input := &organizations.DescribeAccountInput{
		AccountId: aws.String(id),
	}

	output, err := conn.DescribeAccount(input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAccountNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Account == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := aws.StringValue(output.Account.Status); status == organizations.AccountStatusSuspended {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output.Account, nil
}

func FindOrganization(conn *organizations.Organizations) (*organizations.Organization, error) {
	input := &organizations.DescribeOrganizationInput{}

	output, err := conn.DescribeOrganization(input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Organization == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Organization, nil
}

func FindPolicyAttachmentByTwoPartKey(conn *organizations.Organizations, targetID, policyID string) (*organizations.PolicyTargetSummary, error) {
	input := &organizations.ListTargetsForPolicyInput{
		PolicyId: aws.String(policyID),
	}
	var output *organizations.PolicyTargetSummary

	err := conn.ListTargetsForPolicyPages(input, func(page *organizations.ListTargetsForPolicyOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Targets {
			if aws.StringValue(v.TargetId) == targetID {
				output = v
				return true
			}
		}
		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeTargetNotFoundException, organizations.ErrCodePolicyNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{}
	}

	return output, nil
}
