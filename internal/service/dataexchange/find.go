// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func FindDataSetById(ctx context.Context, conn *dataexchange.Client, id string) (*dataexchange.GetDataSetOutput, error) {
	input := &dataexchange.GetDataSetInput{
		DataSetId: aws.String(id),
	}
	output, err := conn.GetDataSet(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindRevisionById(ctx context.Context, conn *dataexchange.Client, dataSetId, revisionId string) (*dataexchange.GetRevisionOutput, error) {
	input := &dataexchange.GetRevisionInput{
		DataSetId:  aws.String(dataSetId),
		RevisionId: aws.String(revisionId),
	}
	output, err := conn.GetRevision(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
