package kendra

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindFaqByID(ctx context.Context, conn *kendra.Client, id, indexId string) (*kendra.DescribeFaqOutput, error) {
	in := &kendra.DescribeFaqInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	}

	out, err := conn.DescribeFaq(ctx, in)

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func FindQuerySuggestionsBlockListByID(ctx context.Context, conn *kendra.Client, id, indexId string) (*kendra.DescribeQuerySuggestionsBlockListOutput, error) {
	in := &kendra.DescribeQuerySuggestionsBlockListInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	}

	out, err := conn.DescribeQuerySuggestionsBlockList(ctx, in)
	if err != nil {
		var resourceNotFoundException *types.ResourceNotFoundException

		if errors.As(err, &resourceNotFoundException) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func FindThesaurusByID(ctx context.Context, conn *kendra.Client, id, indexId string) (*kendra.DescribeThesaurusOutput, error) {
	in := &kendra.DescribeThesaurusInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	}

	out, err := conn.DescribeThesaurus(ctx, in)

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
