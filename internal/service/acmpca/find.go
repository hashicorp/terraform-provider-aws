package acmpca

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindCertificateAuthorityByARN returns the certificate authority corresponding to the specified ARN.
// Returns nil if no certificate authority is found.
func FindCertificateAuthorityByARN(ctx context.Context, conn *acmpca.ACMPCA, arn string) (*acmpca.CertificateAuthority, error) {
	input := &acmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(arn),
	}

	output, err := conn.DescribeCertificateAuthorityWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.CertificateAuthority, nil
}

// FindCertificateAuthorityCertificateByARN returns the certificate for the certificate authority corresponding to the specified ARN.
// Returns a resource.NotFoundError if no certificate authority is found or the certificate authority does not have a certificate assigned.
func FindCertificateAuthorityCertificateByARN(ctx context.Context, conn *acmpca.ACMPCA, arn string) (*acmpca.GetCertificateAuthorityCertificateOutput, error) {
	input := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(arn),
	}

	output, err := conn.GetCertificateAuthorityCertificateWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

func FindPolicyByARN(ctx context.Context, conn *acmpca.ACMPCA, arn string) (string, error) {
	input := &acmpca.GetPolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		return "", &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Policy == nil {
		return "", tfresource.NewEmptyResultError(input)
	}

	return aws.StringValue(output.Policy), nil
}

func FindPermission(ctx context.Context, conn *acmpca.ACMPCA, certificateAuthorityARN, principal, sourceAccount string) (*acmpca.Permission, error) {
	input := &acmpca.ListPermissionsInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}
	var output []*acmpca.Permission

	err := conn.ListPermissionsPagesWithContext(ctx, input, func(page *acmpca.ListPermissionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Permissions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidStateException, "The certificate authority is in the DELETED state") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.StringValue(v.Principal) == principal && (sourceAccount == "" || aws.StringValue(v.SourceAccount) == sourceAccount) {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{LastRequest: input}
}
