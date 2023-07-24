// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				ValidateFunc: validation.StringInSlice(apprunner.ProviderType_Values(), false),
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
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	name := d.Get("connection_name").(string)
	input := &apprunner.CreateConnectionInput{
		ConnectionName: aws.String(name),
		ProviderType:   aws.String(d.Get("provider_type").(string)),
		Tags:           getTagsIn(ctx),
	}

	output, err := conn.CreateConnectionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating App Runner Connection (%s): %s", name, err)
	}

	if output == nil || output.Connection == nil {
		return diag.Errorf("creating App Runner Connection (%s): empty output", name)
	}

	d.SetId(aws.StringValue(output.Connection.ConnectionName))

	return resourceConnectionRead(ctx, d, meta)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	c, err := FindConnectionSummaryByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
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

	arn := aws.StringValue(c.ConnectionArn)

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
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Get("arn").(string)),
	}

	_, err := conn.DeleteConnectionWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("deleting App Runner Connection (%s): %s", d.Id(), err)
	}

	if err := WaitConnectionDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("waiting for App Runner Connection (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
