package directoryservicedata

import "github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"

func RegisterSweepers() {
	awsv2.Register("aws_directoryservicedata_user", sweepUsers)
}
