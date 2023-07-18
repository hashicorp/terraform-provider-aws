// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_athena_named_query")
func ResourceNamedQuery() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNamedQueryCreate,
		ReadWithoutTimeout:   resourceNamedQueryRead,
		DeleteWithoutTimeout: resourceNamedQueryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"query": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"workgroup": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "primary",
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNamedQueryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaConn(ctx)

	input := &athena.CreateNamedQueryInput{
		Database:    aws.String(d.Get("database").(string)),
		Name:        aws.String(d.Get("name").(string)),
		QueryString: aws.String(d.Get("query").(string)),
	}
	if raw, ok := d.GetOk("workgroup"); ok {
		input.WorkGroup = aws.String(raw.(string))
	}
	if raw, ok := d.GetOk("description"); ok {
		input.Description = aws.String(raw.(string))
	}

	resp, err := conn.CreateNamedQueryWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena Named Query (%s): %s", d.Get("name").(string), err)
	}
	d.SetId(aws.StringValue(resp.NamedQueryId))
	return append(diags, resourceNamedQueryRead(ctx, d, meta)...)
}

func resourceNamedQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaConn(ctx)

	input := &athena.GetNamedQueryInput{
		NamedQueryId: aws.String(d.Id()),
	}

	resp, err := conn.GetNamedQueryWithContext(ctx, input)
	if tfawserr.ErrMessageContains(err, athena.ErrCodeInvalidRequestException, d.Id()) && !d.IsNewResource() {
		log.Printf("[WARN] Athena Named Query (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Athena Named Query (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.NamedQuery.Name)
	d.Set("query", resp.NamedQuery.QueryString)
	d.Set("workgroup", resp.NamedQuery.WorkGroup)
	d.Set("database", resp.NamedQuery.Database)
	d.Set("description", resp.NamedQuery.Description)
	return diags
}

func resourceNamedQueryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaConn(ctx)

	input := &athena.DeleteNamedQueryInput{
		NamedQueryId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteNamedQueryWithContext(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena Named Query (%s): %s", d.Id(), err)
	}
	return diags
}
