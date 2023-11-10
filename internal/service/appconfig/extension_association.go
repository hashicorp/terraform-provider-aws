// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"extension_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"resource_arn": {
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
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	in := appconfig.CreateExtensionAssociationInput{
		ExtensionIdentifier: aws.String(d.Get("extension_arn").(string)),
		ResourceIdentifier:  aws.String(d.Get("resource_arn").(string)),
	}

	if v, ok := d.GetOk("parameters"); ok {
		in.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	out, err := conn.CreateExtensionAssociationWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionCreating, ResExtensionAssociation, d.Get("extension_arn").(string), err)
	}

	if out == nil {
		return create.DiagError(names.AppConfig, create.ErrActionCreating, ResExtensionAssociation, d.Get("extension_arn").(string), errors.New("No Extension Association returned with create request."))
	}

	d.SetId(aws.StringValue(out.Id))

	return resourceExtensionAssociationRead(ctx, d, meta)
}

func resourceExtensionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	out, err := FindExtensionAssociationById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.AppConfig, create.ErrActionReading, ResExtensionAssociation, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionReading, ResExtensionAssociation, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("extension_arn", out.ExtensionArn)
	d.Set("parameters", out.Parameters)
	d.Set("resource_arn", out.ResourceArn)
	d.Set("extension_version", out.ExtensionVersionNumber)

	return nil
}

func resourceExtensionAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)
	requestUpdate := false

	in := &appconfig.UpdateExtensionAssociationInput{
		ExtensionAssociationId: aws.String(d.Id()),
	}

	if d.HasChange("parameters") {
		in.Parameters = flex.ExpandStringMap(d.Get("parameters").(map[string]interface{}))
		requestUpdate = true
	}

	if requestUpdate {
		out, err := conn.UpdateExtensionAssociationWithContext(ctx, in)

		if err != nil {
			return create.DiagError(names.AppConfig, create.ErrActionWaitingForUpdate, ResExtensionAssociation, d.Id(), err)
		}

		if out == nil {
			return create.DiagError(names.AppConfig, create.ErrActionWaitingForUpdate, ResExtensionAssociation, d.Id(), errors.New("No ExtensionAssociation returned with update request."))
		}
	}

	return resourceExtensionAssociationRead(ctx, d, meta)
}

func resourceExtensionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	log.Printf("[INFO] Deleting AppConfig Hosted Extension Association: %s", d.Id())
	_, err := conn.DeleteExtensionAssociationWithContext(ctx, &appconfig.DeleteExtensionAssociationInput{
		ExtensionAssociationId: aws.String(d.Id()),
	})

	if err != nil {
		return create.DiagError(names.AppConfig, create.ErrActionDeleting, ResExtensionAssociation, d.Id(), err)
	}

	return nil
}
