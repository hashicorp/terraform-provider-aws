// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_application", name="Application")
// @Tags(identifierAttribute="arn")
func ResourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationCreate,
		ReadWithoutTimeout:   resourceApplicationRead,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	applicationName := d.Get("name").(string)
	input := &appconfig.CreateApplicationInput{
		Name: aws.String(applicationName),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	app, err := conn.CreateApplicationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Application (%s): %s", applicationName, err)
	}

	if app == nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Application (%s): empty response", applicationName)
	}

	d.SetId(aws.StringValue(app.Id))

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	input := &appconfig.GetApplicationInput{
		ApplicationId: aws.String(d.Id()),
	}

	output, err := conn.GetApplicationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appconfig Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting AppConfig Application (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting AppConfig Application (%s): empty response", d.Id())
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s", aws.StringValue(output.Id)),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)
	d.Set("name", output.Name)
	d.Set("description", output.Description)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		updateInput := &appconfig.UpdateApplicationInput{
			ApplicationId: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			updateInput.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			updateInput.Name = aws.String(d.Get("name").(string))
		}

		_, err := conn.UpdateApplicationWithContext(ctx, updateInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Application(%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	log.Printf("[INFO] Deleting AppConfig Application: %s", d.Id())
	_, err := conn.DeleteApplicationWithContext(ctx, &appconfig.DeleteApplicationInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appconfig Application (%s): %s", d.Id(), err)
	}

	return diags
}
