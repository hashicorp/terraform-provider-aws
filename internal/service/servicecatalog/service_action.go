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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_service_action")
func ResourceServiceAction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceActionCreate,
		ReadWithoutTimeout:   resourceServiceActionRead,
		UpdateWithoutTimeout: resourceServiceActionUpdate,
		DeleteWithoutTimeout: resourceServiceActionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ServiceActionReadyTimeout),
			Read:   schema.DefaultTimeout(ServiceActionReadTimeout),
			Update: schema.DefaultTimeout(ServiceActionUpdateTimeout),
			Delete: schema.DefaultTimeout(ServiceActionDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      AcceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
			},
			"definition": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assume_role": { // ServiceActionDefinitionKeyAssumeRole
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrName: { // ServiceActionDefinitionKeyName
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrParameters: { // ServiceActionDefinitionKeyParameters
							Type:             schema.TypeString,
							Optional:         true,
							ValidateFunc:     validation.StringIsJSON,
							DiffSuppressFunc: suppressEquivalentJSONEmptyNilDiffs,
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      servicecatalog.ServiceActionDefinitionTypeSsmAutomation,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(servicecatalog.ServiceActionDefinitionType_Values(), false),
						},
						names.AttrVersion: { // ServiceActionDefinitionKeyVersion
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceServiceActionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.CreateServiceActionInput{
		IdempotencyToken: aws.String(id.UniqueId()),
		Name:             aws.String(d.Get(names.AttrName).(string)),
		Definition:       expandServiceActionDefinition(d.Get("definition").([]interface{})[0].(map[string]interface{})),
		DefinitionType:   aws.String(d.Get("definition.0.type").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	var output *servicecatalog.CreateServiceActionOutput
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.CreateServiceActionWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateServiceActionWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Service Action: %s", err)
	}

	if output == nil || output.ServiceActionDetail == nil || output.ServiceActionDetail.ServiceActionSummary == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Service Action: empty response")
	}

	d.SetId(aws.StringValue(output.ServiceActionDetail.ServiceActionSummary.Id))

	return append(diags, resourceServiceActionRead(ctx, d, meta)...)
}

func resourceServiceActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	output, err := WaitServiceActionReady(ctx, conn, d.Get("accept_language").(string), d.Id(), d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Service Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Service Action (%s): %s", d.Id(), err)
	}

	if output == nil || output.ServiceActionSummary == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Service Action (%s): empty response", d.Id())
	}

	sas := output.ServiceActionSummary

	d.Set(names.AttrDescription, sas.Description)
	d.Set(names.AttrName, sas.Name)

	if output.Definition != nil {
		d.Set("definition", []interface{}{flattenServiceActionDefinition(output.Definition, aws.StringValue(sas.DefinitionType))})
	} else {
		d.Set("definition", nil)
	}

	return diags
}

func resourceServiceActionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.UpdateServiceActionInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChange("accept_language") {
		input.AcceptLanguage = aws.String(d.Get("accept_language").(string))
	}

	if d.HasChange("definition") {
		input.Definition = expandServiceActionDefinition(d.Get("definition").([]interface{})[0].(map[string]interface{}))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange(names.AttrName) {
		input.Name = aws.String(d.Get(names.AttrName).(string))
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		_, err := conn.UpdateServiceActionWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateServiceActionWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Service Action (%s): %s", d.Id(), err)
	}

	return append(diags, resourceServiceActionRead(ctx, d, meta)...)
}

func resourceServiceActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.DeleteServiceActionInput{
		Id: aws.String(d.Id()),
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := conn.DeleteServiceActionWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceInUseException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteServiceActionWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[INFO] Attempted to delete Service Action (%s) but does not exist", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Service Action (%s): %s", d.Id(), err)
	}

	if err := WaitServiceActionDeleted(ctx, conn, d.Get("accept_language").(string), d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Service Action (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func expandServiceActionDefinition(tfMap map[string]interface{}) map[string]*string {
	if tfMap == nil {
		return nil
	}

	apiObject := make(map[string]*string)

	if v, ok := tfMap["assume_role"].(string); ok && v != "" {
		apiObject[servicecatalog.ServiceActionDefinitionKeyAssumeRole] = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject[servicecatalog.ServiceActionDefinitionKeyName] = aws.String(v)
	}

	if v, ok := tfMap[names.AttrParameters].(string); ok && v != "" {
		apiObject[servicecatalog.ServiceActionDefinitionKeyParameters] = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject[servicecatalog.ServiceActionDefinitionKeyVersion] = aws.String(v)
	}

	return apiObject
}

func flattenServiceActionDefinition(apiObject map[string]*string, definitionType string) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v, ok := apiObject[servicecatalog.ServiceActionDefinitionKeyAssumeRole]; ok && v != nil {
		tfMap["assume_role"] = aws.StringValue(v)
	}

	if v, ok := apiObject[servicecatalog.ServiceActionDefinitionKeyName]; ok && v != nil {
		tfMap[names.AttrName] = aws.StringValue(v)
	}

	if v, ok := apiObject[servicecatalog.ServiceActionDefinitionKeyParameters]; ok && v != nil {
		tfMap[names.AttrParameters] = aws.StringValue(v)
	}

	if v, ok := apiObject[servicecatalog.ServiceActionDefinitionKeyVersion]; ok && v != nil {
		tfMap[names.AttrVersion] = aws.StringValue(v)
	}

	if definitionType != "" {
		tfMap[names.AttrType] = definitionType
	}

	return tfMap
}
