package ssm

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAssociationById(ctx context.Context, conn *ssm.SSM, id string) (*ssm.AssociationDescription, error) {
	input := &ssm.DescribeAssociationInput{
		AssociationId: aws.String(id),
	}

	output, err := conn.DescribeAssociationWithContext(ctx, input)
	if tfawserr.ErrCodeContains(err, ssm.ErrCodeAssociationDoesNotExist) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AssociationDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AssociationDescription, nil
}

// FindDocumentByName returns the Document corresponding to the specified name.
func FindDocumentByName(ctx context.Context, conn *ssm.SSM, name string) (*ssm.DocumentDescription, error) {
	input := &ssm.DescribeDocumentInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeDocumentWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.Document == nil {
		return nil, fmt.Errorf("error describing SSM Document (%s): empty result", name)
	}

	doc := output.Document

	if aws.StringValue(doc.Status) == ssm.DocumentStatusFailed {
		return nil, fmt.Errorf("Document is in a failed state: %s", aws.StringValue(doc.StatusInformation))
	}

	return output.Document, nil
}

// FindPatchGroup returns matching SSM Patch Group by Patch Group and BaselineId.
func FindPatchGroup(ctx context.Context, conn *ssm.SSM, patchGroup, baselineId string) (*ssm.PatchGroupPatchBaselineMapping, error) {
	input := &ssm.DescribePatchGroupsInput{}
	var result *ssm.PatchGroupPatchBaselineMapping

	err := conn.DescribePatchGroupsPagesWithContext(ctx, input, func(page *ssm.DescribePatchGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, mapping := range page.Mappings {
			if mapping == nil {
				continue
			}

			if aws.StringValue(mapping.PatchGroup) == patchGroup {
				if mapping.BaselineIdentity != nil && aws.StringValue(mapping.BaselineIdentity.BaselineId) == baselineId {
					result = mapping
					return false
				}
			}
		}

		return !lastPage
	})

	return result, err
}

func FindServiceSettingByID(ctx context.Context, conn *ssm.SSM, id string) (*ssm.ServiceSetting, error) {
	input := &ssm.GetServiceSettingInput{
		SettingId: aws.String(id),
	}

	output, err := conn.GetServiceSettingWithContext(ctx, input)

	if tfawserr.ErrCodeContains(err, ssm.ErrCodeServiceSettingNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServiceSetting == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ServiceSetting, nil
}
