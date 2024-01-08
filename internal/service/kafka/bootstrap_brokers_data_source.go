// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_msk_bootstrap_brokers", name="Bootstrap Brokers")
func dataSourceBootstrapBrokers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBootstrapBrokersRead,

		Schema: map[string]*schema.Schema{
			"bootstrap_broker_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_public_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_public_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_public_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_vpc_connectivity_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_vpc_connectivity_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_vpc_connectivity_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBootstrapBrokersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn(ctx)

	clusterARN := d.Get("cluster_arn").(string)
	output, err := findBootstrapBrokersByARN(ctx, conn, clusterARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster (%s) bootstrap brokers: %s", clusterARN, err)
	}

	d.Set("bootstrap_broker_string", output.BootstrapBrokerString)
	d.Set("bootstrap_broker_string_tls", output.BootstrapBrokerStringTls)
	d.Set("bootstrap_broker_string_sasl_scram", output.BootstrapBrokerStringSaslScram)
	d.Set("bootstrap_broker_string_sasl_iam", output.BootstrapBrokerStringSaslIam)
	d.Set("bootstrap_broker_string_public_tls", output.BootstrapBrokerStringPublicTls)
	d.Set("bootstrap_broker_string_public_sasl_scram", output.BootstrapBrokerStringPublicSaslScram)
	d.Set("bootstrap_broker_string_public_sasl_iam", output.BootstrapBrokerStringPublicSaslIam)
	d.Set("bootstrap_broker_string_vpc_connectivity_tls", output.BootstrapBrokerStringVpcConnectivityTls)
	d.Set("bootstrap_broker_string_vpc_connectivity_sasl_scram", output.BootstrapBrokerStringVpcConnectivitySaslScram)
	d.Set("bootstrap_broker_string_vpc_connectivity_sasl_iam", output.BootstrapBrokerStringVpcConnectivitySaslIam)

	return diags
}
