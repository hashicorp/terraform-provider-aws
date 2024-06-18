// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_mskconnect_connector")
func DataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectorRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaConnectConn(ctx)

	name := d.Get(names.AttrName)
	var output []*kafkaconnect.ConnectorSummary

	err := conn.ListConnectorsPagesWithContext(ctx, &kafkaconnect.ListConnectorsInput{}, func(page *kafkaconnect.ListConnectorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connectors {
			if aws.StringValue(v.ConnectorName) == name {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing MSK Connect Connectors: %s", err)
	}

	if len(output) == 0 || output[0] == nil {
		err = tfresource.NewEmptyResultError(name)
	} else if count := len(output); count > 1 {
		err = tfresource.NewTooManyResultsError(count, name)
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MSK Connect Connector", err))
	}

	connector := output[0]

	d.SetId(aws.StringValue(connector.ConnectorArn))

	d.Set(names.AttrARN, connector.ConnectorArn)
	d.Set(names.AttrDescription, connector.ConnectorDescription)
	d.Set(names.AttrName, connector.ConnectorName)
	d.Set(names.AttrVersion, connector.CurrentVersion)

	return diags
}
