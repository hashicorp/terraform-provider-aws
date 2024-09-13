// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssm_document", name="Document")
func dataSourceDocument() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataDocumentRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_format": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DocumentFormatJson,
				ValidateDiagFunc: enum.Validate[awstypes.DocumentFormat](),
			},
			"document_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ssm.GetDocumentInput{
		DocumentFormat: awstypes.DocumentFormat(d.Get("document_format").(string)),
		Name:           aws.String(name),
	}

	if v, ok := d.GetOk("document_version"); ok {
		input.DocumentVersion = aws.String(v.(string))
	}

	output, err := conn.GetDocument(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Document (%s) content: %s", name, err)
	}

	d.SetId(aws.ToString(output.Name))

	if !strings.HasPrefix(name, "AWS-") {
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "ssm",
			Region:    meta.(*conns.AWSClient).Region,
			AccountID: meta.(*conns.AWSClient).AccountID,
			Resource:  "document/" + name,
		}.String()
		d.Set(names.AttrARN, arn)
	} else {
		d.Set(names.AttrARN, name)
	}
	d.Set(names.AttrContent, output.Content)
	d.Set("document_format", output.DocumentFormat)
	d.Set("document_type", output.DocumentType)
	d.Set("document_version", output.DocumentVersion)
	d.Set(names.AttrName, output.Name)

	return diags
}
