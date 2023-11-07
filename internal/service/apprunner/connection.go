// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apprunner_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connection_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"provider_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(flattenProviderTypeValues(types.ProviderType("").Values()), false),
				ForceNew:     true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	name := d.Get("connection_name").(string)
	input := &apprunner.CreateConnectionInput{
		ConnectionName: aws.String(name),
		ProviderType:   types.ProviderType(d.Get("provider_type").(string)),
		Tags:           getTagsIn(ctx),
	}

	output, err := conn.CreateConnection(ctx, input)

	if err != nil {
		return diag.Errorf("creating App Runner Connection (%s): %s", name, err)
	}

	if output == nil || output.Connection == nil {
		return diag.Errorf("creating App Runner Connection (%s): empty output", name)
	}

	d.SetId(aws.ToString(output.Connection.ConnectionName))

	return resourceConnectionRead(ctx, d, meta)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	c, err := FindConnectionsummaryByName(ctx, conn, d.Id())

	if !d.IsNewResource() && errs.IsA[*types.ResourceNotFoundException](err) {
		log.Printf("[WARN] App Runner Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading App Runner Connection (%s): %s", d.Id(), err)
	}

	if c == nil {
		if d.IsNewResource() {
			return diag.Errorf("reading App Runner Connection (%s): empty output after creation", d.Id())
		}
		log.Printf("[WARN] App Runner Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.ToString(c.ConnectionArn)

	d.Set("arn", arn)
	d.Set("connection_name", c.ConnectionName)
	d.Set("provider_type", c.ProviderType)
	d.Set("status", c.Status)

	return nil
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceConnectionRead(ctx, d, meta)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	input := &apprunner.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Get("arn").(string)),
	}

	_, err := conn.DeleteConnection(ctx, input)

	if err != nil {
		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil
		}
		return diag.Errorf("deleting App Runner Connection (%s): %s", d.Id(), err)
	}

	if err := WaitConnectionDeleted(ctx, conn, d.Id()); err != nil {
		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil
		}
		return diag.Errorf("waiting for App Runner Connection (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func flattenProviderTypeValues(t []types.ProviderType) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}
