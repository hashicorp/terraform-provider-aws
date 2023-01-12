package s3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/endpoints"
)

// See https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_website_region_endpoints.
var hostedZoneIDsMap = map[string]string{
	endpoints.AfSouth1RegionID:     "Z83WF9RJE8B12",
	endpoints.ApEast1RegionID:      "ZNB98KWMFR0R6",
	endpoints.ApNortheast1RegionID: "Z2M4EHUR26P7ZW",
	endpoints.ApNortheast2RegionID: "Z3W03O7B5YMIYP",
	endpoints.ApNortheast3RegionID: "Z2YQB5RD63NC85",
	endpoints.ApSouth1RegionID:     "Z11RGJOFQNVJUP",
	endpoints.ApSoutheast1RegionID: "Z3O0J2DXBE1FTB",
	endpoints.ApSoutheast2RegionID: "Z1WCIGYICN2BYD",
	endpoints.ApSoutheast3RegionID: "Z01613992JD795ZI93075",
	endpoints.CaCentral1RegionID:   "Z1QDHH18159H29",
	endpoints.CnNorthwest1RegionID: "Z282HJ1KT0DH03",
	endpoints.EuCentral1RegionID:   "Z21DNDUVLTQW6Q",
	endpoints.EuCentral2RegionID:   "Z030506016YDQGETNASS",
	endpoints.EuNorth1RegionID:     "Z3BAZG2TWCNX0D",
	endpoints.EuSouth1RegionID:     "Z30OZKI7KPW7MI",
	endpoints.EuSouth2RegionID:     "Z0081959F7139GRJC19J",
	endpoints.EuWest1RegionID:      "Z1BKCTXD74EZPE",
	endpoints.EuWest2RegionID:      "Z3GKZC51ZF0DB4",
	endpoints.EuWest3RegionID:      "Z3R1K369G5AVDG",
	endpoints.MeCentral1RegionID:   "Z06143092I8HRXZRUZROF",
	endpoints.MeSouth1RegionID:     "Z1MPMWCPA7YB62",
	endpoints.SaEast1RegionID:      "Z7KQH4QJS55SO",
	endpoints.UsEast1RegionID:      "Z3AQBSTGFYJSTF",
	endpoints.UsEast2RegionID:      "Z2O1EMRO9K5GLX",
	endpoints.UsGovEast1RegionID:   "Z2NIFVYYW2VKV1",
	endpoints.UsGovWest1RegionID:   "Z31GFT0UA1I2HV",
	endpoints.UsWest1RegionID:      "Z2F56UZL2M1ACD",
	endpoints.UsWest2RegionID:      "Z3BJ6K6RIION7M",
}

// Returns the hosted zone ID for an S3 website endpoint region. This can be
// used as input to the aws_route53_record resource's zone_id argument.
func HostedZoneIDForRegion(region string) (string, error) {
	if v, ok := hostedZoneIDsMap[region]; ok {
		return v, nil
	}
	return "", fmt.Errorf("S3 hosted zone ID not found for region: %s", region)
}
