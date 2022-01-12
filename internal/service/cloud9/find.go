package cloud9

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindEnvironmentByID(conn *cloud9.Cloud9, id string) (*cloud9.Environment, error) {
	input := &cloud9.DescribeEnvironmentsInput{
		EnvironmentIds: []*string{aws.String(id)},
	}
	out, err := conn.DescribeEnvironments(input)

	if tfawserr.ErrCodeEquals(err, cloud9.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	envs := out.Environments

	if len(envs) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	env := envs[0]

	if env == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return env, nil
}

func FindEnvironmentMembershipByID(conn *cloud9.Cloud9, envId, userArn string) (*cloud9.EnvironmentMember, error) {
	input := &cloud9.DescribeEnvironmentMembershipsInput{
		EnvironmentId: aws.String(envId),
		UserArn:       aws.String(userArn),
	}
	out, err := conn.DescribeEnvironmentMemberships(input)

	if tfawserr.ErrCodeEquals(err, cloud9.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	envs := out.Memberships

	if len(envs) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	env := envs[0]

	if env == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return env, nil
}
