package drs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/drs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/drs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindReplicationConfigurationTemplateByID(ctx context.Context, conn *drs.Client, id string) (*awstypes.ReplicationConfigurationTemplate, error) {
	//make string list from id
	templateIdList := []string{}
	templateIdList = append(templateIdList, id)

	input := &awstypes.DescribeReplicationConfigurationTemplatesInput{
		ReplicationConfigurationTemplateIDs: templateIdList,
	}

	output, err := conn.DescribeReplicationConfigurationTemplates(input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
