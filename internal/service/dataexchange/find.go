// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindDataSetById(ctx context.Context, conn *dataexchange.DataExchange, id string) (*dataexchange.GetDataSetOutput, error) {
	input := &dataexchange.GetDataSetInput{
		DataSetId: aws.String(id),
	}
	output, err := conn.GetDataSetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dataexchange.ErrCodeResourceNotFoundException) {
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

func FindRevisionById(ctx context.Context, conn *dataexchange.DataExchange, dataSetId, revisionId string) (*dataexchange.GetRevisionOutput, error) {
	input := &dataexchange.GetRevisionInput{
		DataSetId:  aws.String(dataSetId),
		RevisionId: aws.String(revisionId),
	}
	output, err := conn.GetRevisionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dataexchange.ErrCodeResourceNotFoundException) {
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
