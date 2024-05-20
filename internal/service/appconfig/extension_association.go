// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResExtensionAssociation = "ExtensionAssociation"
)

// @SDKResource("aws_appconfig_extension_association")
func ResourceExtensionAssociation() *schema.Resource {
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"extension_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceExtensionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	in := appconfig.CreateExtensionAssociationInput{
		ExtensionIdentifier: aws.String(d.Get("extension_arn").(string)),
		ResourceIdentifier:  aws.String(d.Get(names.AttrResourceARN).(string)),
	}

	if v, ok := d.GetOk(names.AttrParameters); ok {
		in.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	out, err := conn.CreateExtensionAssociation(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionCreating, ResExtensionAssociation, d.Get("extension_arn").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionCreating, ResExtensionAssociation, d.Get("extension_arn").(string), errors.New("No Extension Association returned with create request."))
	}

	d.SetId(aws.ToString(out.Id))

	return append(diags, resourceExtensionAssociationRead(ctx, d, meta)...)
}

func resourceExtensionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	out, err := FindExtensionAssociationById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.AppConfig, create.ErrActionReading, ResExtensionAssociation, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, ResExtensionAssociation, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set("extension_arn", out.ExtensionArn)
	d.Set(names.AttrParameters, out.Parameters)
	d.Set(names.AttrResourceARN, out.ResourceArn)
	d.Set("extension_version", out.ExtensionVersionNumber)

	return diags
}

func resourceExtensionAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)
	requestUpdate := false

	in := &appconfig.UpdateExtensionAssociationInput{
		ExtensionAssociationId: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrParameters) {
		in.Parameters = flex.ExpandStringValueMap(d.Get(names.AttrParameters).(map[string]interface{}))
		requestUpdate = true
	}

	if requestUpdate {
		out, err := conn.UpdateExtensionAssociation(ctx, in)

		if err != nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionWaitingForUpdate, ResExtensionAssociation, d.Id(), err)
		}

		if out == nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionWaitingForUpdate, ResExtensionAssociation, d.Id(), errors.New("No ExtensionAssociation returned with update request."))
		}
	}

	return append(diags, resourceExtensionAssociationRead(ctx, d, meta)...)
}

func resourceExtensionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	log.Printf("[INFO] Deleting AppConfig Hosted Extension Association: %s", d.Id())
	_, err := conn.DeleteExtensionAssociation(ctx, &appconfig.DeleteExtensionAssociationInput{
		ExtensionAssociationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.AppConfig, create.ErrActionDeleting, ResExtensionAssociation, d.Id(), err)
	}

	return diags
}
