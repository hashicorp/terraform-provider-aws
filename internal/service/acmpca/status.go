package acmpca

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/acmpca/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	certificateAuthorityStatusNotFound = "NotFound"
	certificateAuthorityStatusUnknown  = "Unknown"
)

// statusCertificateAuthority fetches the Deployment and its Status
func statusCertificateAuthority(conn *acmpca.ACMPCA, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		certificateAuthority, err := finder.FindCertificateAuthorityByARN(conn, arn)

		if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
			return nil, certificateAuthorityStatusNotFound, nil
		}

		if err != nil {
			return nil, certificateAuthorityStatusUnknown, err
		}

		if certificateAuthority == nil {
			return nil, certificateAuthorityStatusNotFound, nil
		}

		return certificateAuthority, aws.StringValue(certificateAuthority.Status), nil
	}
}
