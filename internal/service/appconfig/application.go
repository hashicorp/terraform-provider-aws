// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_application", name="Application")
// @Tags(identifierAttribute="arn")
func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationCreate,
		ReadWithoutTimeout:   resourceApplicationRead,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationName := d.Get(names.AttrName).(string)
	input := appconfig.CreateApplicationInput{
		Name: aws.String(applicationName),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateApplication(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Application (%s): %s", applicationName, err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	output, err := findApplicationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppConfig Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Application (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, applicationARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrName, output.Name)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := appconfig.UpdateApplicationInput{
			ApplicationId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err := conn.UpdateApplication(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Application (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	log.Printf("[INFO] Deleting AppConfig Application: %s", d.Id())
	input := appconfig.DeleteApplicationInput{
		ApplicationId: aws.String(d.Id()),
	}
	_, err := conn.DeleteApplication(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppConfig Application (%s): %s", d.Id(), err)
	}

	return diags
}

func findApplicationByID(ctx context.Context, conn *appconfig.Client, id string) (*appconfig.GetApplicationOutput, error) {
	input := appconfig.GetApplicationInput{
		ApplicationId: aws.String(id),
	}

	return findApplication(ctx, conn, &input)
}

func findApplication(ctx context.Context, conn *appconfig.Client, input *appconfig.GetApplicationInput) (*appconfig.GetApplicationOutput, error) {
	output, err := conn.GetApplication(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func applicationARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, "appconfig", "application/"+id)
}
