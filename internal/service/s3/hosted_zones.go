// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-aws/names"
)

// See https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_website_region_endpoints.
var hostedZoneIDsMap = map[string]string{
	names.AFSouth1RegionID:     "Z83WF9RJE8B12",
	names.APEast1RegionID:      "ZNB98KWMFR0R6",
	names.APNortheast1RegionID: "Z2M4EHUR26P7ZW",
	names.APNortheast2RegionID: "Z3W03O7B5YMIYP",
	names.APNortheast3RegionID: "Z2YQB5RD63NC85",
	names.APSouth1RegionID:     "Z11RGJOFQNVJUP",
	names.APSouth2RegionID:     "Z02976202B4EZMXIPMXF7",
	names.APSoutheast1RegionID: "Z3O0J2DXBE1FTB",
	names.APSoutheast2RegionID: "Z1WCIGYICN2BYD",
	names.APSoutheast3RegionID: "Z01846753K324LI26A3VV",
	names.APSoutheast4RegionID: "Z0312387243XT5FE14WFO",
	names.CACentral1RegionID:   "Z1QDHH18159H29",
	names.CAWest1RegionID:      "Z03565811Z33SLEZTHOUL",
	names.CNNorth1RegionID:     "Z5CN8UMXT92WN",
	names.CNNorthwest1RegionID: "Z282HJ1KT0DH03",
	names.EUCentral1RegionID:   "Z21DNDUVLTQW6Q",
	names.EUCentral2RegionID:   "Z030506016YDQGETNASS",
	names.EUNorth1RegionID:     "Z3BAZG2TWCNX0D",
	names.EUSouth1RegionID:     "Z30OZKI7KPW7MI",
	names.EUSouth2RegionID:     "Z0081959F7139GRJC19J",
	names.EUWest1RegionID:      "Z1BKCTXD74EZPE",
	names.EUWest2RegionID:      "Z3GKZC51ZF0DB4",
	names.EUWest3RegionID:      "Z3R1K369G5AVDG",
	names.ILCentral1RegionID:   "Z09640613K4A3MN55U7GU",
	names.MECentral1RegionID:   "Z06143092I8HRXZRUZROF",
	names.MESouth1RegionID:     "Z1MPMWCPA7YB62",
	names.SAEast1RegionID:      "Z7KQH4QJS55SO",
	names.USEast1RegionID:      "Z3AQBSTGFYJSTF",
	names.USEast2RegionID:      "Z2O1EMRO9K5GLX",
	names.USGovEast1RegionID:   "Z2NIFVYYW2VKV1",
	names.USGovWest1RegionID:   "Z31GFT0UA1I2HV",
	names.USWest1RegionID:      "Z2F56UZL2M1ACD",
	names.USWest2RegionID:      "Z3BJ6K6RIION7M",
}

// hostedZoneIDForRegion returns the Route 53 hosted zone ID for an S3 website endpoint Region.
// This can be used as input to the aws_route53_record resource's zone_id argument.
func hostedZoneIDForRegion(region string) (string, error) {
	if v, ok := hostedZoneIDsMap[region]; ok {
		return v, nil
	}
	return "", fmt.Errorf("S3 website Route 53 hosted zone ID not found for Region (%s)", region)
}
