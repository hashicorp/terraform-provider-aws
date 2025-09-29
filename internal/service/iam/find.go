// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindSigningCertificate(ctx context.Context, conn *iam.Client, userName, certId string) (*awstypes.SigningCertificate, error) {
	input := &iam.ListSigningCertificatesInput{
		UserName: aws.String(userName),
	}

	output, err := conn.ListSigningCertificates(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.Certificates) == 0 {
		return nil, tfresource.NewEmptyResultError(output)
	}

	var cert awstypes.SigningCertificate

	for _, crt := range output.Certificates {
		if aws.ToString(crt.UserName) == userName &&
			aws.ToString(crt.CertificateId) == certId {
			cert = crt
			break
		}
	}

	if reflect.ValueOf(cert).IsZero() {
		return nil, tfresource.NewEmptyResultError(cert)
	}

	return &cert, nil
}
