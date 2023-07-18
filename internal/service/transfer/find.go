// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAccessByTwoPartKey(ctx context.Context, conn *transfer.Transfer, serverID, externalID string) (*transfer.DescribedAccess, error) {
	input := &transfer.DescribeAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	output, err := conn.DescribeAccessWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Access == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Access, nil
}

func FindAgreementByTwoPartKey(ctx context.Context, conn *transfer.Transfer, serverID, agreementID string) (*transfer.DescribedAgreement, error) {
	input := &transfer.DescribeAgreementInput{
		AgreementId: aws.String(agreementID),
		ServerId:    aws.String(serverID),
	}

	output, err := conn.DescribeAgreementWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Agreement == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Agreement, nil
}

func FindCertificateByID(ctx context.Context, conn *transfer.Transfer, id string) (*transfer.DescribedCertificate, error) {
	input := &transfer.DescribeCertificateInput{
		CertificateId: aws.String(id),
	}

	output, err := conn.DescribeCertificateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Certificate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Certificate, nil
}

func FindConnectorByID(ctx context.Context, conn *transfer.Transfer, id string) (*transfer.DescribedConnector, error) {
	input := &transfer.DescribeConnectorInput{
		ConnectorId: aws.String(id),
	}

	output, err := conn.DescribeConnectorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connector == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Connector, nil
}

func FindProfileByID(ctx context.Context, conn *transfer.Transfer, id string) (*transfer.DescribedProfile, error) {
	input := &transfer.DescribeProfileInput{
		ProfileId: aws.String(id),
	}

	output, err := conn.DescribeProfileWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Profile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Profile, nil
}

func FindServerByID(ctx context.Context, conn *transfer.Transfer, id string) (*transfer.DescribedServer, error) {
	input := &transfer.DescribeServerInput{
		ServerId: aws.String(id),
	}

	output, err := conn.DescribeServerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Server == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Server, nil
}

func FindWorkflowByID(ctx context.Context, conn *transfer.Transfer, id string) (*transfer.DescribedWorkflow, error) {
	input := &transfer.DescribeWorkflowInput{
		WorkflowId: aws.String(id),
	}

	output, err := conn.DescribeWorkflowWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workflow == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workflow, nil
}
