// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"

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

// @SDKResource("aws_servicecatalog_budget_resource_association")
func ResourceBudgetResourceAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBudgetResourceAssociationCreate,
		ReadWithoutTimeout:   resourceBudgetResourceAssociationRead,
		DeleteWithoutTimeout: resourceBudgetResourceAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(BudgetResourceAssociationReadyTimeout),
			Read:   schema.DefaultTimeout(BudgetResourceAssociationReadTimeout),
			Delete: schema.DefaultTimeout(BudgetResourceAssociationDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"budget_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBudgetResourceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.AssociateBudgetWithResourceInput{
		BudgetName: aws.String(d.Get("budget_name").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
	}

	var output *servicecatalog.AssociateBudgetWithResourceOutput
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.AssociateBudgetWithResourceWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociateBudgetWithResourceWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating Service Catalog Budget with Resource: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Budget Resource Association: empty response")
	}

	d.SetId(BudgetResourceAssociationID(d.Get("budget_name").(string), d.Get(names.AttrResourceID).(string)))

	return append(diags, resourceBudgetResourceAssociationRead(ctx, d, meta)...)
}

func resourceBudgetResourceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	budgetName, resourceID, err := BudgetResourceAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	output, err := WaitBudgetResourceAssociationReady(ctx, conn, budgetName, resourceID, d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Budget Resource Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Budget Resource Association (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Budget Resource Association (%s): empty response", d.Id())
	}

	d.Set(names.AttrResourceID, resourceID)
	d.Set("budget_name", output.BudgetName)

	return diags
}

func resourceBudgetResourceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	budgetName, resourceID, err := BudgetResourceAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "could not parse ID (%s): %s", d.Id(), err)
	}

	input := &servicecatalog.DisassociateBudgetFromResourceInput{
		ResourceId: aws.String(resourceID),
		BudgetName: aws.String(budgetName),
	}

	_, err = conn.DisassociateBudgetFromResourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Service Catalog Budget from Resource (%s): %s", d.Id(), err)
	}

	err = WaitBudgetResourceAssociationDeleted(ctx, conn, budgetName, resourceID, d.Timeout(schema.TimeoutDelete))

	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Budget Resource Disassociation (%s): %s", d.Id(), err)
	}

	return diags
}
