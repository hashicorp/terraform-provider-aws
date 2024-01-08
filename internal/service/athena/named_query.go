// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_athena_named_query")
func resourceNamedQuery() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNamedQueryCreate,
		ReadWithoutTimeout:   resourceNamedQueryRead,
		DeleteWithoutTimeout: resourceNamedQueryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func resourceNamedQueryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	name := d.Get("name").(string)
	input := &athena.CreateNamedQueryInput{
		Database:    aws.String(d.Get("database").(string)),
		Name:        aws.String(name),
		QueryString: aws.String(d.Get("query").(string)),
	}

	if v, ok := d.GetOk("workgroup"); ok {
		input.WorkGroup = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateNamedQuery(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena Named Query (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.NamedQueryId))

	return append(diags, resourceNamedQueryRead(ctx, d, meta)...)
}

func resourceNamedQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	namedQuery, err := findNamedQueryByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Athena Named Query (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Athena Named Query (%s): %s", d.Id(), err)
	}

	d.Set("database", namedQuery.Database)
	d.Set("description", namedQuery.Description)
	d.Set("name", namedQuery.Name)
	d.Set("query", namedQuery.QueryString)
	d.Set("workgroup", namedQuery.WorkGroup)

	return diags
}

func resourceNamedQueryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	log.Printf("[INFO] Deleting Athena Named Query: %s", d.Id())
	_, err := conn.DeleteNamedQuery(ctx, &athena.DeleteNamedQueryInput{
		NamedQueryId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena Named Query (%s): %s", d.Id(), err)
	}

	return diags
}

func findNamedQueryByID(ctx context.Context, conn *athena.Client, id string) (*types.NamedQuery, error) {
	input := &athena.GetNamedQueryInput{
		NamedQueryId: aws.String(id),
	}

	output, err := conn.GetNamedQuery(ctx, input)

	if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.NamedQuery == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.NamedQuery, nil
}
