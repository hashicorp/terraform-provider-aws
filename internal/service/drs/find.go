package drs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/drs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindReplicationConfigurationTemplateByID(conn *drs.Drs, id string) (*drs.ReplicationConfigurationTemplate, error) {
	//make string list from id
	templateIdList := []*string{}
	templateIdList = append(templateIdList, aws.String(id))

	input := &drs.DescribeReplicationConfigurationTemplatesInput{
		ReplicationConfigurationTemplateIDs: templateIdList,
	}

	output, err := conn.DescribeReplicationConfigurationTemplates(input)

	if tfawserr.ErrCodeEquals(err, drs.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Items == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Items[0], nil
}
