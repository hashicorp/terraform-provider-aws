package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func HealthCheckByID(conn *route53.Route53, id string) (*route53.HealthCheck, error) {
	input := &route53.GetHealthCheckInput{
		HealthCheckId: aws.String(id),
	}

	output, err := conn.GetHealthCheck(input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHealthCheck) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || output.HealthCheck == nil || output.HealthCheck.HealthCheckConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.HealthCheck, nil
}

func HostedZoneDnssec(conn *route53.Route53, hostedZoneID string) (*route53.GetDNSSECOutput, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.GetDNSSEC(input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func KeySigningKey(conn *route53.Route53, hostedZoneID string, name string) (*route53.KeySigningKey, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	var result *route53.KeySigningKey

	output, err := conn.GetDNSSEC(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, keySigningKey := range output.KeySigningKeys {
		if keySigningKey == nil {
			continue
		}

		if aws.StringValue(keySigningKey.Name) == name {
			result = keySigningKey
			break
		}
	}

	return result, err
}

func KeySigningKeyByResourceID(conn *route53.Route53, resourceID string) (*route53.KeySigningKey, error) {
	hostedZoneID, name, err := tfroute53.KeySigningKeyParseResourceID(resourceID)

	if err != nil {
		return nil, fmt.Errorf("error parsing Route 53 Key Signing Key (%s) identifier: %w", resourceID, err)
	}

	return KeySigningKey(conn, hostedZoneID, name)
}
