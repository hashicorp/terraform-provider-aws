// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicecatalog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_provisioned_product", name="Provisioned Product")
// @Tags
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/servicecatalog/types;awstypes;awstypes.ProvisionedProductDetail",importIgnore="accept_language;ignore_errors;provisioning_artifact_name;provisioning_parameters;retain_physical_resources", skipEmptyTags=true, noRemoveTags=true)
// @Testing(tagsIdentifierAttribute="id", tagsResourceType="Provisioned Product")
// @Testing(tagsUpdateGetTagsIn=true)
func resourceProvisionedProduct() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProvisionedProductCreate,
		ReadWithoutTimeout:   resourceProvisionedProductRead,
		UpdateWithoutTimeout: resourceProvisionedProductUpdate,
		DeleteWithoutTimeout: resourceProvisionedProductDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      acceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(acceptLanguage_Values(), false),
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

		CustomizeDiff: refreshOutputsDiff,
	}
}

func refreshOutputsDiff(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	if diff.HasChanges("provisioning_parameters", "provisioning_artifact_id", "provisioning_artifact_name") {
		if err := diff.SetNewComputed("outputs"); err != nil {
			return err
		}
	}

	return nil
}

func resourceProvisionedProductCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := servicecatalog.ProvisionProductInput{
		ProvisionToken:         aws.String(sdkid.UniqueId()),
		ProvisionedProductName: aws.String(name),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_arns"); ok && len(v.([]any)) > 0 {
		input.NotificationArns = flex.ExpandStringValueList(v.([]any))
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

	if v, ok := d.GetOk("provisioning_parameters"); ok && len(v.([]any)) > 0 {
		input.ProvisioningParameters = expandProvisioningParameters(v.([]any))
	}

	if v, ok := d.GetOk("stack_set_provisioning_preferences"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ProvisioningPreferences = expandProvisioningPreferences(v.([]any)[0].(map[string]any))
	}

	output, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutCreate),
		func(ctx context.Context) (*servicecatalog.ProvisionProductOutput, error) {
			return conn.ProvisionProduct(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidParametersException](err, "profile does not exist") {
				return true, err
			}

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "provisioning Service Catalog Product (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.RecordDetail.ProvisionedProductId))

	if _, err := waitProvisionedProductReady(ctx, conn, d.Id(), d.Get("accept_language").(string), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Provisioned Product (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceProvisionedProductRead(ctx, d, meta)...)
}

func resourceProvisionedProductRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	// There are two API operations for getting information about provisioned products:
	// 1. DescribeProvisionedProduct (used in WaitProvisionedProductReady) and
	// 2. DescribeRecord (used in WaitRecordReady)

	// They provide some overlapping information. Most of the unique information available from
	// DescribeRecord is available in the data source aws_servicecatalog_record.

	acceptLanguage := acceptLanguageEnglish
	if v, ok := d.GetOk("accept_language"); ok {
		acceptLanguage = v.(string)
	}

	outputDPP, err := findProvisionedProductByTwoPartKey(ctx, conn, d.Id(), acceptLanguage)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Service Catalog Provisioned Product (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Catalog Provisioned Product (%s): %s", d.Id(), err)
	}

	detail := outputDPP.ProvisionedProductDetail
	d.Set(names.AttrARN, detail.Arn)
	d.Set("cloudwatch_dashboard_names", flattenCloudWatchDashboards(outputDPP.CloudWatchDashboards))
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

	recordIdToUse := aws.ToString(detail.LastProvisioningRecordId)
	if detail.Status == awstypes.ProvisionedProductStatusTainted && detail.LastSuccessfulProvisioningRecordId != nil {
		// For TAINTED resources, we need to get artifact details from the last successful
		// record, as the last provisioned record may have failed.
		recordIdToUse = aws.ToString(detail.LastSuccessfulProvisioningRecordId)
		log.Printf("[DEBUG] Service Catalog Provisioned Product (%s) is TAINTED, using last successful record %s for parameter values", d.Id(), recordIdToUse)
	}
	outputDR, err := findRecordByTwoPartKey(ctx, conn, recordIdToUse, acceptLanguage)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Service Catalog Provisioned Product (%s) Record (%s) not found, unable to set tags", d.Id(), recordIdToUse)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Catalog Provisioned Product (%s) Record (%s): %s", d.Id(), recordIdToUse, err)
	}

	// To enable debugging of potential issues, log as a warning instead of exiting prematurely.
	// For example, errors can be present after a failed version update, and the stack rolled back
	// to the current version.
	if v := outputDR.RecordDetail.RecordErrors; len(v) > 0 {
		var errs []error

		for _, err := range v {
			errs = append(errs, fmt.Errorf("%s: %s", aws.ToString(err.Code), aws.ToString(err.Description)))
		}

		log.Printf("[WARN] Errors found when describing Service Catalog Provisioned Product (%s) Record (%s): %s", d.Id(), aws.ToString(detail.LastProvisioningRecordId), errors.Join(errs...))
	}

	if err := d.Set("outputs", flattenRecordOutputs(outputDR.RecordOutputs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting outputs: %s", err)
	}

	d.Set("path_id", outputDR.RecordDetail.PathId)
	d.Set("provisioning_artifact_id", outputDR.RecordDetail.ProvisioningArtifactId)

	setTagsOut(ctx, svcTags(recordKeyValueTags(ctx, outputDR.RecordDetail.RecordTags)))

	return diags
}

func resourceProvisionedProductUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	input := servicecatalog.UpdateProvisionedProductInput{
		ProvisionedProductId: aws.String(d.Id()),
		UpdateToken:          aws.String(sdkid.UniqueId()),
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

	if v, ok := d.GetOk("provisioning_parameters"); ok && len(v.([]any)) > 0 {
		input.ProvisioningParameters = expandUpdateProvisioningParameters(v.([]any))
	}

	if v, ok := d.GetOk("stack_set_provisioning_preferences"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ProvisioningPreferences = expandUpdateProvisioningPreferences(v.([]any)[0].(map[string]any))
	}

	// Send tags each time the resource is updated. This is necessary to automatically apply tags
	// to provisioned AWS objects during update if the tags don't change.
	input.Tags = getTagsIn(ctx)

	_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutUpdate),
		func(ctx context.Context) (any, error) {
			return conn.UpdateProvisionedProduct(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidParametersException](err, "profile does not exist") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Provisioned Product (%s): %s", d.Id(), err)
	}

	if _, err := waitProvisionedProductReady(ctx, conn, d.Id(), d.Get("accept_language").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
		if failureErr, ok := errs.As[*provisionedProductFailureError](err); ok {
			log.Printf("[WARN] Service Catalog Provisioned Product (%s) update failed with status %s, refreshing state", d.Id(), failureErr.Status)
			refreshDiags := resourceProvisionedProductRead(ctx, d, meta)
			if refreshDiags.HasError() {
				// If refresh fails, return both errors
				return append(refreshDiags, sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Provisioned Product (%s) update: %s", d.Id(), err)...)
			}

			if d.HasChange("provisioning_parameters") {
				// If parameters were changed, rollback to previous values.
				//
				// The read APIs used to refresh state above do not return parameter values, and therefore
				// will not reflect that the planned updates did not take effect. Explicitly rolling back
				// ensures the planned parameter changes are attempted again on a subsequent apply.
				oldParams, _ := d.GetChange("provisioning_parameters")
				d.Set("provisioning_parameters", oldParams)
			}
		}

		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Provisioned Product (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceProvisionedProductRead(ctx, d, meta)...)
}

func resourceProvisionedProductDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	input := servicecatalog.TerminateProvisionedProductInput{
		TerminateToken:       aws.String(sdkid.UniqueId()),
		ProvisionedProductId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ignore_errors"); ok {
		input.IgnoreErrors = v.(bool)
	}

	if v, ok := d.GetOk("retain_physical_resources"); ok {
		input.RetainPhysicalResources = v.(bool)
	}

	_, err := conn.TerminateProvisionedProduct(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "terminating Service Catalog Provisioned Product (%s): %s", d.Id(), err)
	}

	_, err = waitProvisionedProductTerminated(ctx, conn, d.Id(), d.Get("accept_language").(string), d.Timeout(schema.TimeoutDelete))

	if errs.IsA[*provisionedProductFailureError](err) {
		input.IgnoreErrors = true

		_, err = conn.TerminateProvisionedProduct(ctx, &input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "terminating Service Catalog Provisioned Product (%s): %s", d.Id(), err)
		}

		_, err = waitProvisionedProductTerminated(ctx, conn, d.Id(), d.Get("accept_language").(string), d.Timeout(schema.TimeoutDelete))
	}

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Provisioned Product (%s) terminate: %s", d.Id(), err)
	}

	return diags
}

func findProvisionedProductByTwoPartKey(ctx context.Context, conn *servicecatalog.Client, id, acceptLanguage string) (*servicecatalog.DescribeProvisionedProductOutput, error) {
	input := servicecatalog.DescribeProvisionedProductInput{
		Id: aws.String(id),
	}
	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	return findProvisionedProduct(ctx, conn, &input)
}

func findProvisionedProduct(ctx context.Context, conn *servicecatalog.Client, input *servicecatalog.DescribeProvisionedProductInput) (*servicecatalog.DescribeProvisionedProductOutput, error) {
	output, err := conn.DescribeProvisionedProduct(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ProvisionedProductDetail == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findRecordByTwoPartKey(ctx context.Context, conn *servicecatalog.Client, id, acceptLanguage string) (*servicecatalog.DescribeRecordOutput, error) {
	input := servicecatalog.DescribeRecordInput{
		Id: aws.String(id),
	}
	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	return findRecord(ctx, conn, &input)
}

func findRecord(ctx context.Context, conn *servicecatalog.Client, input *servicecatalog.DescribeRecordInput) (*servicecatalog.DescribeRecordOutput, error) {
	var output *servicecatalog.DescribeRecordOutput

	for {
		page, err := conn.DescribeRecord(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		if page == nil {
			break
		}

		if output == nil {
			output = page
		} else {
			output.RecordOutputs = append(output.RecordOutputs, page.RecordOutputs...)
		}

		nextPageToken := aws.ToString(page.NextPageToken)
		if nextPageToken == "" {
			break
		}
		input.PageToken = aws.String(nextPageToken)
	}

	if output == nil || output.RecordDetail == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusProvisionedProduct(conn *servicecatalog.Client, id, acceptLanguage string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findProvisionedProductByTwoPartKey(ctx, conn, id, acceptLanguage)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ProvisionedProductDetail.Status), err
	}
}

func waitProvisionedProductReady(ctx context.Context, conn *servicecatalog.Client, id, acceptLanguage string, timeout time.Duration) (*servicecatalog.DescribeProvisionedProductOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProvisionedProductStatusUnderChange, awstypes.ProvisionedProductStatusPlanInProgress),
		Target:                    enum.Slice(awstypes.ProvisionedProductStatusAvailable),
		Refresh:                   statusProvisionedProduct(conn, id, acceptLanguage),
		Timeout:                   timeout,
		ContinuousTargetOccurence: continuousTargetOccurrence,
		NotFoundChecks:            notFoundChecks,
		MinTimeout:                minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisionedProductOutput); ok {
		if detail := output.ProvisionedProductDetail; detail != nil {
			if errs.IsA[*retry.UnexpectedStateError](err) {
				// The statuses `ERROR` and `TAINTED` are equivalent: the application of the requested change has failed.
				// The difference is that, in the case of `TAINTED`, there is a previous version to roll back to.
				if status := detail.Status; status == awstypes.ProvisionedProductStatusError || status == awstypes.ProvisionedProductStatusTainted {
					// Create a custom error type that signals state refresh is needed
					return output, &provisionedProductFailureError{
						StatusMessage: aws.ToString(detail.StatusMessage),
						Status:        status,
					}
				}
			}
		}
		return output, err
	}

	return nil, err
}

func waitProvisionedProductTerminated(ctx context.Context, conn *servicecatalog.Client, id, acceptLanguage string, timeout time.Duration) (*servicecatalog.DescribeProvisionedProductOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ProvisionedProductStatusAvailable,
			awstypes.ProvisionedProductStatusUnderChange,
		),
		Target:  []string{},
		Refresh: statusProvisionedProduct(conn, id, acceptLanguage),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisionedProductOutput); ok {
		if detail := output.ProvisionedProductDetail; detail != nil {
			if errs.IsA[*retry.UnexpectedStateError](err) {
				// If the status is `TAINTED`, we can retry with `IgnoreErrors`
				if status := detail.Status; status == awstypes.ProvisionedProductStatusTainted {
					// Create a custom error type that signals state refresh is needed
					return output, &provisionedProductFailureError{
						StatusMessage: aws.ToString(detail.StatusMessage),
						Status:        status,
					}
				}
			}
		}
		return output, err
	}

	return nil, err
}

// provisionedProductFailureError represents a provisioned product operation failure
// that requires state refresh to recover from inconsistent state.
type provisionedProductFailureError struct {
	StatusMessage string
	Status        awstypes.ProvisionedProductStatus
}

func (e *provisionedProductFailureError) Error() string {
	return e.StatusMessage
}

func expandProvisioningParameter(tfMap map[string]any) awstypes.ProvisioningParameter {
	apiObject := awstypes.ProvisioningParameter{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandProvisioningParameters(tfList []any) []awstypes.ProvisioningParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ProvisioningParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandProvisioningParameter(tfMap))
	}

	return apiObjects
}

func expandProvisioningPreferences(tfMap map[string]any) *awstypes.ProvisioningPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ProvisioningPreferences{}

	if v, ok := tfMap["accounts"].([]any); ok && len(v) > 0 {
		apiObject.StackSetAccounts = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["failure_tolerance_count"].(int); ok && v != 0 {
		apiObject.StackSetFailureToleranceCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["failure_tolerance_percentage"].(int); ok && v != 0 {
		apiObject.StackSetFailureTolerancePercentage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_concurrency_count"].(int); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_concurrency_percentage"].(int); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyPercentage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["regions"].([]any); ok && len(v) > 0 {
		apiObject.StackSetRegions = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandUpdateProvisioningParameter(tfMap map[string]any) awstypes.UpdateProvisioningParameter {
	apiObject := awstypes.UpdateProvisioningParameter{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["use_previous_value"].(bool); ok && v {
		apiObject.UsePreviousValue = v
	}

	if v, ok := tfMap[names.AttrValue].(string); ok {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandUpdateProvisioningParameters(tfList []any) []awstypes.UpdateProvisioningParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.UpdateProvisioningParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandUpdateProvisioningParameter(tfMap))
	}

	return apiObjects
}

func expandUpdateProvisioningPreferences(tfMap map[string]any) *awstypes.UpdateProvisioningPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.UpdateProvisioningPreferences{}

	if v, ok := tfMap["accounts"].([]any); ok && len(v) > 0 {
		apiObject.StackSetAccounts = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["failure_tolerance_count"].(int); ok && v != 0 {
		apiObject.StackSetFailureToleranceCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["failure_tolerance_percentage"].(int); ok && v != 0 {
		apiObject.StackSetFailureTolerancePercentage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_concurrency_count"].(int); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_concurrency_percentage"].(int); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyPercentage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["regions"].([]any); ok && len(v) > 0 {
		apiObject.StackSetRegions = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func flattenCloudWatchDashboards(apiObjects []awstypes.CloudWatchDashboard) []*string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, apiObject.Name)
	}

	return tfList
}

func flattenRecordOutputs(apiObjects []awstypes.RecordOutput) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		m := make(map[string]any)

		if apiObject.Description != nil {
			m[names.AttrDescription] = aws.ToString(apiObject.Description)
		}

		if apiObject.OutputKey != nil {
			m[names.AttrKey] = aws.ToString(apiObject.OutputKey)
		}

		if apiObject.OutputValue != nil {
			m[names.AttrValue] = aws.ToString(apiObject.OutputValue)
		}

		tfList = append(tfList, m)
	}

	return tfList
}
