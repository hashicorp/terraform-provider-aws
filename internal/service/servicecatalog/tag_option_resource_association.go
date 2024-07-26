// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_tag_option_resource_association")
func ResourceTagOptionResourceAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTagOptionResourceAssociationCreate,
		ReadWithoutTimeout:   resourceTagOptionResourceAssociationRead,
		DeleteWithoutTimeout: resourceTagOptionResourceAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(TagOptionResourceAssociationReadyTimeout),
			Read:   schema.DefaultTimeout(TagOptionResourceAssociationReadTimeout),
			Delete: schema.DefaultTimeout(TagOptionResourceAssociationDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tag_option_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTagOptionResourceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.AssociateTagOptionWithResourceInput{
		ResourceId:  aws.String(d.Get(names.AttrResourceID).(string)),
		TagOptionId: aws.String(d.Get("tag_option_id").(string)),
	}

	var output *servicecatalog.AssociateTagOptionWithResourceOutput
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.AssociateTagOptionWithResourceWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociateTagOptionWithResourceWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating Service Catalog Tag Option with Resource: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Tag Option Resource Association: empty response")
	}

	d.SetId(TagOptionResourceAssociationID(d.Get("tag_option_id").(string), d.Get(names.AttrResourceID).(string)))

	return append(diags, resourceTagOptionResourceAssociationRead(ctx, d, meta)...)
}

func resourceTagOptionResourceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	tagOptionID, resourceID, err := TagOptionResourceAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	output, err := WaitTagOptionResourceAssociationReady(ctx, conn, tagOptionID, resourceID, d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Tag Option Resource Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Tag Option Resource Association (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Tag Option Resource Association (%s): empty response", d.Id())
	}

	if output.CreatedTime != nil {
		d.Set("resource_created_time", output.CreatedTime.Format(time.RFC3339))
	}

	d.Set(names.AttrResourceARN, output.ARN)
	d.Set("resource_description", output.Description)
	d.Set(names.AttrResourceID, output.Id)
	d.Set("resource_name", output.Name)
	d.Set("tag_option_id", tagOptionID)

	return diags
}

func resourceTagOptionResourceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	tagOptionID, resourceID, err := TagOptionResourceAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	input := &servicecatalog.DisassociateTagOptionFromResourceInput{
		ResourceId:  aws.String(resourceID),
		TagOptionId: aws.String(tagOptionID),
	}

	_, err = conn.DisassociateTagOptionFromResourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Service Catalog Tag Option from Resource (%s): %s", d.Id(), err)
	}

	err = WaitTagOptionResourceAssociationDeleted(ctx, conn, tagOptionID, resourceID, d.Timeout(schema.TimeoutDelete))

	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Tag Option Resource Disassociation (%s): %s", d.Id(), err)
	}

	return diags
}
