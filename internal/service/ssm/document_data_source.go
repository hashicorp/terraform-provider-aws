// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_ssm_document")
func DataSourceDocument() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataDocumentRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_format": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ssm.DocumentFormatJson,
				ValidateFunc: validation.StringInSlice(ssm.DocumentFormat_Values(), false),
			},
			"document_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn(ctx)

	name := d.Get("name").(string)
	input := &ssm.GetDocumentInput{
		DocumentFormat: aws.String(d.Get("document_format").(string)),
		Name:           aws.String(name),
	}

	if v, ok := d.GetOk("document_version"); ok {
		input.DocumentVersion = aws.String(v.(string))
	}

	output, err := conn.GetDocumentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Document (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Name))

	if !strings.HasPrefix(name, "AWS-") {
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "ssm",
			Region:    meta.(*conns.AWSClient).Region,
			AccountID: meta.(*conns.AWSClient).AccountID,
			Resource:  fmt.Sprintf("document/%s", name),
		}.String()
		d.Set("arn", arn)
	} else {
		d.Set("arn", name)
	}
	d.Set("content", output.Content)
	d.Set("document_format", output.DocumentFormat)
	d.Set("document_type", output.DocumentType)
	d.Set("document_version", output.DocumentVersion)
	d.Set("name", output.Name)

	return diags
}
