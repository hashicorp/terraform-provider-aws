package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsRoute53ZoneImportState(
	d *schema.ResourceData,
	meta interface{}) ([]*schema.ResourceData, error) {

	results := make([]*schema.ResourceData, 1, 1)
	results[0] = d

	conn := meta.(*AWSClient).r53conn
	zone, err := conn.GetHostedZone(&route53.GetHostedZoneInput{Id: aws.String(d.Id())})
	if err != nil {
		return results, err
	}

	if len(zone.VPCs) > 1 {
		for index, vpc := range zone.VPCs {
			if index == 0 {
				d.Set("vpc_region", vpc.VPCRegion)
				d.Set("vpc_id", vpc.VPCId)
				continue
			}

			zoneAssociation := resourceAwsRoute53ZoneAssociation()
			associationData := zoneAssociation.Data(nil)
			associationData.SetType("aws_route53_zone_association")
			associationData.SetId(fmt.Sprintf("%s:%s", d.Id(), *vpc.VPCId))
			associationData.Set("vpc_region", vpc.VPCRegion)
			associationData.Set("vpc_id", vpc.VPCId)
			associationData.Set("zone_id", d.Id())
			results = append(results, associationData)
		}

	}

	return results, nil
}
