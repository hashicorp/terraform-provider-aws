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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_extension_association", name="Extension Association")
func resourceExtensionAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceExtensionAssociationCreate,
		ReadWithoutTimeout:   resourceExtensionAssociationRead,
		UpdateWithoutTimeout: resourceExtensionAssociationUpdate,
		DeleteWithoutTimeout: resourceExtensionAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"extension_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"extension_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceExtensionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	input := appconfig.CreateExtensionAssociationInput{
		ExtensionIdentifier: aws.String(d.Get("extension_arn").(string)),
		ResourceIdentifier:  aws.String(d.Get(names.AttrResourceARN).(string)),
	}

	if v, ok := d.GetOk(names.AttrParameters); ok {
		input.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
	}

	output, err := conn.CreateExtensionAssociation(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Extension Association: %s", err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceExtensionAssociationRead(ctx, d, meta)...)
}

func resourceExtensionAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	output, err := findExtensionAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppConfig Extension Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Extension Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set("extension_arn", output.ExtensionArn)
	d.Set("extension_version", output.ExtensionVersionNumber)
	d.Set(names.AttrParameters, output.Parameters)
	d.Set(names.AttrResourceARN, output.ResourceArn)

	return diags
}

func resourceExtensionAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	input := appconfig.UpdateExtensionAssociationInput{
		ExtensionAssociationId: aws.String(d.Id()),
		Parameters:             flex.ExpandStringValueMap(d.Get(names.AttrParameters).(map[string]any)),
	}

	_, err := conn.UpdateExtensionAssociation(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppConfig Extension Association (%s): %s", d.Id(), err)
	}

	return append(diags, resourceExtensionAssociationRead(ctx, d, meta)...)
}

func resourceExtensionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	log.Printf("[INFO] Deleting AppConfig Extension Association: %s", d.Id())
	input := appconfig.DeleteExtensionAssociationInput{
		ExtensionAssociationId: aws.String(d.Id()),
	}
	_, err := conn.DeleteExtensionAssociation(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppConfig Extension Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findExtensionAssociationByID(ctx context.Context, conn *appconfig.Client, id string) (*appconfig.GetExtensionAssociationOutput, error) {
	input := appconfig.GetExtensionAssociationInput{
		ExtensionAssociationId: aws.String(id),
	}

	return findExtensionAssociation(ctx, conn, &input)
}

func findExtensionAssociation(ctx context.Context, conn *appconfig.Client, input *appconfig.GetExtensionAssociationInput) (*appconfig.GetExtensionAssociationOutput, error) {
	output, err := conn.GetExtensionAssociation(ctx, input)

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
