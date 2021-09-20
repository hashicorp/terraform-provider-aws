package route53

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/route53/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func statusChangeInfo(conn *route53.Route53, changeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &route53.GetChangeInput{
			Id: aws.String(changeID),
		}

		output, err := conn.GetChange(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.ChangeInfo == nil {
			return nil, "", nil
		}

		return output.ChangeInfo, aws.StringValue(output.ChangeInfo.Status), nil
	}
}

func statusHostedZoneDNSSEC(conn *route53.Route53, hostedZoneID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		hostedZoneDnssec, err := finder.FindHostedZoneDNSSEC(conn, hostedZoneID)

		if err != nil {
			return nil, "", err
		}

		if hostedZoneDnssec == nil || hostedZoneDnssec.Status == nil {
			return nil, "", nil
		}

		return hostedZoneDnssec.Status, aws.StringValue(hostedZoneDnssec.Status.ServeSignature), nil
	}
}

func statusKeySigningKey(conn *route53.Route53, hostedZoneID string, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		keySigningKey, err := finder.FindKeySigningKey(conn, hostedZoneID, name)

		if err != nil {
			return nil, "", err
		}

		if keySigningKey == nil {
			return nil, "", nil
		}

		return keySigningKey, aws.StringValue(keySigningKey.Status), nil
	}
}
