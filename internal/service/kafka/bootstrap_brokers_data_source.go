// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_msk_bootstrap_brokers", name="Bootstrap Brokers")
func dataSourceBootstrapBrokers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBootstrapBrokersRead,

		Schema: map[string]*schema.Schema{
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_public_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_public_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_public_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_vpc_connectivity_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_vpc_connectivity_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_vpc_connectivity_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func dataSourceBootstrapBrokersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	clusterARN := d.Get("cluster_arn").(string)
	output, err := findBootstrapBrokersByARN(ctx, conn, clusterARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster (%s) bootstrap brokers: %s", clusterARN, err)
	}

	d.SetId(clusterARN)
	d.Set("bootstrap_brokers", SortEndpointsString(aws.ToString(output.BootstrapBrokerString)))
	d.Set("bootstrap_brokers_public_sasl_iam", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringPublicSaslIam)))
	d.Set("bootstrap_brokers_public_sasl_scram", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringPublicSaslScram)))
	d.Set("bootstrap_brokers_public_tls", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringPublicTls)))
	d.Set("bootstrap_brokers_sasl_iam", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringSaslIam)))
	d.Set("bootstrap_brokers_sasl_scram", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringSaslScram)))
	d.Set("bootstrap_brokers_tls", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringTls)))
	d.Set("bootstrap_brokers_vpc_connectivity_sasl_iam", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringVpcConnectivitySaslIam)))
	d.Set("bootstrap_brokers_vpc_connectivity_sasl_scram", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringVpcConnectivitySaslScram)))
	d.Set("bootstrap_brokers_vpc_connectivity_tls", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringVpcConnectivityTls)))

	return diags
}
