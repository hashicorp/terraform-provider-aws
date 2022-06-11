package lightsail

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCertificateByName(ctx context.Context, conn *lightsail.Lightsail, name string) (*lightsail.Certificate, error) {
	in := &lightsail.GetCertificatesInput{
		CertificateName: aws.String(name),
	}

	out, err := conn.GetCertificatesWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Certificates) == 0 || out.Certificates[0] == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Certificates[0].CertificateDetail, nil
}
