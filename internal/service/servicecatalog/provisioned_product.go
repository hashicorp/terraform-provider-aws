// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_provisioned_product", name="Provisioned Product")
// @Tags
// @Testing(existsType="github.com/aws/aws-sdk-go/service/servicecatalog;servicecatalog.ProvisionedProductDetail",importIgnore="accept_language;ignore_errors;provisioning_artifact_name;provisioning_parameters;retain_physical_resources", skipEmptyTags=true, noRemoveTags=true)
func ResourceProvisionedProduct() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProvisionedProductCreate,
		ReadWithoutTimeout:   resourceProvisionedProductRead,
		UpdateWithoutTimeout: resourceProvisionedProductUpdate,
		DeleteWithoutTimeout: resourceProvisionedProductDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ProvisionedProductReadyTimeout),
			Read:   schema.DefaultTimeout(ProvisionedProductReadTimeout),
			Update: schema.DefaultTimeout(ProvisionedProductUpdateTimeout),
			Delete: schema.DefaultTimeout(ProvisionedProductDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudwatch_dashboard_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ignore_errors": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"last_provisioning_record_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_record_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_successful_provisioning_record_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"notification_arns": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"outputs": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"path_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ConflictsWith: []string{
					"path_name",
				},
			},
			"path_name": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"path_id",
				},
			},
			"product_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ExactlyOneOf: []string{
					"product_id",
					"product_name",
				},
			},
			"product_name": {
				Type:     schema.TypeString,
				Optional: true,
				ExactlyOneOf: []string{
					"product_id",
					"product_name",
				},
			},
			"provisioning_artifact_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ExactlyOneOf: []string{
					"provisioning_artifact_id",
					"provisioning_artifact_name",
				},
			},
			"provisioning_artifact_name": {
				Type:     schema.TypeString,
				Optional: true,
				ExactlyOneOf: []string{
					"provisioning_artifact_id",
					"provisioning_artifact_name",
				},
			},
			"provisioning_parameters": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						"use_previous_value": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"retain_physical_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stack_set_provisioning_preferences": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accounts": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"failure_tolerance_count": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"stack_set_provisioning_preferences.0.failure_tolerance_count",
								"stack_set_provisioning_preferences.0.failure_tolerance_percentage",
							},
						},
						"failure_tolerance_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"stack_set_provisioning_preferences.0.failure_tolerance_count",
								"stack_set_provisioning_preferences.0.failure_tolerance_percentage",
							},
						},
						"max_concurrency_count": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"stack_set_provisioning_preferences.0.max_concurrency_count",
								"stack_set_provisioning_preferences.0.max_concurrency_percentage",
							},
						},
						"max_concurrency_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"stack_set_provisioning_preferences.0.max_concurrency_count",
								"stack_set_provisioning_preferences.0.max_concurrency_percentage",
							},
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusMessage: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.All(
			refreshOutputsDiff,
			verify.SetTagsDiff,
		),
	}
}

func refreshOutputsDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if diff.HasChanges("provisioning_parameters", "provisioning_artifact_id", "provisioning_artifact_name") {
		if err := diff.SetNewComputed("outputs"); err != nil {
			return err
		}
	}

	return nil
}

func resourceProvisionedProductCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.ProvisionProductInput{
		ProvisionToken:         aws.String(id.UniqueId()),
		ProvisionedProductName: aws.String(d.Get(names.AttrName).(string)),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_arns"); ok && len(v.([]interface{})) > 0 {
		input.NotificationArns = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("path_id"); ok {
		input.PathId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("path_name"); ok {
		input.PathName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_id"); ok {
		input.ProductId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_name"); ok {
		input.ProductName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_artifact_id"); ok {
		input.ProvisioningArtifactId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_artifact_name"); ok {
		input.ProvisioningArtifactName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_parameters"); ok && len(v.([]interface{})) > 0 {
		input.ProvisioningParameters = expandProvisioningParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("stack_set_provisioning_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ProvisioningPreferences = expandProvisioningPreferences(v.([]interface{})[0].(map[string]interface{}))
	}

	var output *servicecatalog.ProvisionProductOutput

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.ProvisionProductWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.ProvisionProductWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "provisioning Service Catalog Product: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "provisioning Service Catalog Product: empty response")
	}

	if output.RecordDetail == nil {
		return sdkdiag.AppendErrorf(diags, "provisioning Service Catalog Product: no product view detail or summary")
	}

	d.SetId(aws.StringValue(output.RecordDetail.ProvisionedProductId))

	if _, err := WaitProvisionedProductReady(ctx, conn, d.Get("accept_language").(string), d.Id(), "", d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Provisioned Product (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceProvisionedProductRead(ctx, d, meta)...)
}

func resourceProvisionedProductRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	// There are two API operations for getting information about provisioned products:
	// 1. DescribeProvisionedProduct (used in WaitProvisionedProductReady) and
	// 2. DescribeRecord (used in WaitRecordReady)

	// They provide some overlapping information. Most of the unique information available from
	// DescribeRecord is available in the data source aws_servicecatalog_record.

	acceptLanguage := AcceptLanguageEnglish

	if v, ok := d.GetOk("accept_language"); ok {
		acceptLanguage = v.(string)
	}

	input := &servicecatalog.DescribeProvisionedProductInput{
		Id:             aws.String(d.Id()),
		AcceptLanguage: aws.String(acceptLanguage),
	}

	output, err := conn.DescribeProvisionedProductWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Provisioned Product (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Provisioned Product (%s): %s", d.Id(), err)
	}

	if output == nil || output.ProvisionedProductDetail == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Provisioned Product (%s): empty response", d.Id())
	}

	detail := output.ProvisionedProductDetail

	d.Set(names.AttrARN, detail.Arn)
	d.Set("cloudwatch_dashboard_names", aws.StringValueSlice(flattenCloudWatchDashboards(output.CloudWatchDashboards)))

	if detail.CreatedTime != nil {
		d.Set(names.AttrCreatedTime, detail.CreatedTime.Format(time.RFC3339))
	} else {
		d.Set(names.AttrCreatedTime, nil)
	}

	d.Set("last_provisioning_record_id", detail.LastProvisioningRecordId)
	d.Set("last_record_id", detail.LastRecordId)
	d.Set("last_successful_provisioning_record_id", detail.LastSuccessfulProvisioningRecordId)
	d.Set("launch_role_arn", detail.LaunchRoleArn)
	d.Set(names.AttrName, detail.Name)
	d.Set("product_id", detail.ProductId)
	d.Set("provisioning_artifact_id", detail.ProvisioningArtifactId)
	d.Set(names.AttrStatus, detail.Status)
	d.Set(names.AttrStatusMessage, detail.StatusMessage)
	d.Set(names.AttrType, detail.Type)

	// Previously, we waited for the record to only return a target state of 'SUCCEEDED' or 'AVAILABLE'
	// but this can interfere complete reads of this resource when an error occurs after initial creation
	// or after an invalid update that returns a 'FAILED' record state. Thus, waiters are now present in the CREATE and UPDATE methods of this resource instead.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/24574#issuecomment-1126339193
	recordInput := &servicecatalog.DescribeRecordInput{
		Id:             detail.LastProvisioningRecordId,
		AcceptLanguage: aws.String(acceptLanguage),
	}

	recordOutput, err := conn.DescribeRecordWithContext(ctx, recordInput)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Provisioned Product (%s) Record (%s) not found, unable to set tags", d.Id(), aws.StringValue(detail.LastProvisioningRecordId))
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Provisioned Product (%s) Record (%s): %s", d.Id(), aws.StringValue(detail.LastProvisioningRecordId), err)
	}

	if recordOutput == nil || recordOutput.RecordDetail == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Provisioned Product (%s) Record (%s): empty response", d.Id(), aws.StringValue(detail.LastProvisioningRecordId))
	}

	// To enable debugging of potential v, log as a warning
	// instead of exiting prematurely with an error, e.g.
	// v can be present after update to a new version failed and the stack
	// rolled back to the current version.
	if v := recordOutput.RecordDetail.RecordErrors; len(v) > 0 {
		var errs []error

		for _, err := range v {
			errs = append(errs, fmt.Errorf("%s: %s", aws.StringValue(err.Code), aws.StringValue(err.Description)))
		}

		log.Printf("[WARN] Errors found when describing Service Catalog Provisioned Product (%s) Record (%s): %s", d.Id(), aws.StringValue(detail.LastProvisioningRecordId), errors.Join(errs...))
	}

	if err := d.Set("outputs", flattenRecordOutputs(recordOutput.RecordOutputs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting outputs: %s", err)
	}

	d.Set("path_id", recordOutput.RecordDetail.PathId)

	setTagsOut(ctx, Tags(recordKeyValueTags(ctx, recordOutput.RecordDetail.RecordTags)))

	return diags
}

func resourceProvisionedProductUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.UpdateProvisionedProductInput{
		UpdateToken:          aws.String(id.UniqueId()),
		ProvisionedProductId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("path_id"); ok {
		input.PathId = aws.String(v.(string))
	} else if v, ok := d.GetOk("path_name"); ok {
		input.PathName = aws.String(v.(string))
	}

	// check product_name first. product_id is optional/computed and will always be
	// set by the time update is called
	if v, ok := d.GetOk("product_name"); ok {
		input.ProductName = aws.String(v.(string))
	} else if v, ok := d.GetOk("product_id"); ok {
		input.ProductId = aws.String(v.(string))
	}

	// check provisioning_artifact_name first. provisioning_artrifact_id is optional/computed
	// and will always be set by the time update is called
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/26271
	if v, ok := d.GetOk("provisioning_artifact_name"); ok {
		input.ProvisioningArtifactName = aws.String(v.(string))
	} else if v, ok := d.GetOk("provisioning_artifact_id"); ok {
		input.ProvisioningArtifactId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_parameters"); ok && len(v.([]interface{})) > 0 {
		input.ProvisioningParameters = expandUpdateProvisioningParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("stack_set_provisioning_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ProvisioningPreferences = expandUpdateProvisioningPreferences(v.([]interface{})[0].(map[string]interface{}))
	}

	// Send tags each time the resource is updated. This is necessary to automatically apply tags
	// to provisioned AWS objects during update if the tags don't change.
	input.Tags = getTagsIn(ctx)

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		_, err := conn.UpdateProvisionedProductWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateProvisionedProductWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Provisioned Product (%s): %s", d.Id(), err)
	}

	if _, err := WaitProvisionedProductReady(ctx, conn, d.Get("accept_language").(string), d.Id(), "", d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Provisioned Product (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceProvisionedProductRead(ctx, d, meta)...)
}

func resourceProvisionedProductDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.TerminateProvisionedProductInput{
		TerminateToken:       aws.String(id.UniqueId()),
		ProvisionedProductId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ignore_errors"); ok {
		input.IgnoreErrors = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("retain_physical_resources"); ok {
		input.RetainPhysicalResources = aws.Bool(v.(bool))
	}

	_, err := conn.TerminateProvisionedProductWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "terminating Service Catalog Provisioned Product (%s): %s", d.Id(), err)
	}

	err = WaitProvisionedProductTerminated(ctx, conn, d.Get("accept_language").(string), d.Id(), "", d.Timeout(schema.TimeoutDelete))

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Provisioned Product (%s) to be terminated: %s", d.Id(), err)
	}

	return diags
}

func expandProvisioningParameter(tfMap map[string]interface{}) *servicecatalog.ProvisioningParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.ProvisioningParameter{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandProvisioningParameters(tfList []interface{}) []*servicecatalog.ProvisioningParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*servicecatalog.ProvisioningParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandProvisioningParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandProvisioningPreferences(tfMap map[string]interface{}) *servicecatalog.ProvisioningPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.ProvisioningPreferences{}

	if v, ok := tfMap["accounts"].([]interface{}); ok && len(v) > 0 {
		apiObject.StackSetAccounts = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["failure_tolerance_count"].(int); ok && v != 0 {
		apiObject.StackSetFailureToleranceCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["failure_tolerance_percentage"].(int); ok && v != 0 {
		apiObject.StackSetFailureTolerancePercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max_concurrency_count"].(int); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max_concurrency_percentage"].(int); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyPercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["regions"].([]interface{}); ok && len(v) > 0 {
		apiObject.StackSetRegions = flex.ExpandStringList(v)
	}

	return apiObject
}

func expandUpdateProvisioningParameter(tfMap map[string]interface{}) *servicecatalog.UpdateProvisioningParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.UpdateProvisioningParameter{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["use_previous_value"].(bool); ok && v {
		apiObject.UsePreviousValue = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandUpdateProvisioningParameters(tfList []interface{}) []*servicecatalog.UpdateProvisioningParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*servicecatalog.UpdateProvisioningParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandUpdateProvisioningParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandUpdateProvisioningPreferences(tfMap map[string]interface{}) *servicecatalog.UpdateProvisioningPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.UpdateProvisioningPreferences{}

	if v, ok := tfMap["accounts"].([]interface{}); ok && len(v) > 0 {
		apiObject.StackSetAccounts = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["failure_tolerance_count"].(int); ok && v != 0 {
		apiObject.StackSetFailureToleranceCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["failure_tolerance_percentage"].(int); ok && v != 0 {
		apiObject.StackSetFailureTolerancePercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max_concurrency_count"].(int); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max_concurrency_percentage"].(int); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyPercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["regions"].([]interface{}); ok && len(v) > 0 {
		apiObject.StackSetRegions = flex.ExpandStringList(v)
	}

	return apiObject
}

func flattenCloudWatchDashboards(apiObjects []*servicecatalog.CloudWatchDashboard) []*string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, apiObject.Name)
	}

	return tfList
}

func flattenRecordOutputs(apiObjects []*servicecatalog.RecordOutput) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		m := make(map[string]interface{})

		if apiObject.Description != nil {
			m[names.AttrDescription] = aws.StringValue(apiObject.Description)
		}

		if apiObject.OutputKey != nil {
			m[names.AttrKey] = aws.StringValue(apiObject.OutputKey)
		}

		if apiObject.OutputValue != nil {
			m[names.AttrValue] = aws.StringValue(apiObject.OutputValue)
		}

		tfList = append(tfList, m)
	}

	return tfList
}
