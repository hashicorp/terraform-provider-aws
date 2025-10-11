package workmail

import "github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"

func RegisterSweepers() {
	awsv2.Register("aws_workmail_organization", sweepOrganizations)
}
