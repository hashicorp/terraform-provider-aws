package connectcases

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connectcases"
	"github.com/aws/aws-sdk-go-v2/service/connectcases/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findConnectCasesDomainById(ctx context.Context, conn *connectcases.Client, id string) (*connectcases.GetDomainOutput, error) {
	input := &connectcases.GetDomainInput{
		DomainId: aws.String(id),
	}

	output, err := conn.GetDomain(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
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

func findFieldByDomainAndID(ctx context.Context, conn *connectcases.Client, domainId, id string) (*types.FieldSummary, error) {
	input := &connectcases.ListFieldsInput{
		DomainId: aws.String(domainId),
	}

	output, err := conn.ListFields(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
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

	for _, field := range output.Fields {
		if field.FieldId == nil {
			continue
		}
		if aws.ToString(field.FieldId) == id {
			return &field, nil
		}
	}

	return nil, nil
}

func findRelatedItemByID(ctx context.Context, conn *connectcases.Client, caseID, domainID, id string) (*types.SearchRelatedItemsResponseItem, error) {
	input := &connectcases.SearchRelatedItemsInput{
		CaseId:   aws.String(caseID),
		DomainId: aws.String(domainID),
	}

	output, err := conn.SearchRelatedItems(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
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

	for _, relatedItem := range output.RelatedItems {
		if relatedItem.RelatedItemId == nil {
			continue
		}
		if aws.ToString(relatedItem.RelatedItemId) == id {
			return relatedItem, nil
		}
	}

	return nil, nil
}
