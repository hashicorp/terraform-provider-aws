// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_association", name="Association")
// @Tags(identifierAttribute="id", resourceType="Association")
func resourceAssociation() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceAssociationCreate,
		ReadWithoutTimeout:   resourceAssociationRead,
		UpdateWithoutTimeout: resourceAssociationUpdate,
		DeleteWithoutTimeout: resourceAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		MigrateState:  associationMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"apply_only_at_cron_interval": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAssociationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{3,128}$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
				),
			},
			"automation_target_parameter_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"compliance_severity": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ComplianceSeverity](),
			},
			"document_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([$]LATEST|[$]DEFAULT|^[1-9][0-9]*$)$`), ""),
			},
			names.AttrInstanceID: {
				Type:       schema.TypeString,
				ForceNew:   true,
				Optional:   true,
				Deprecated: "use 'targets' argument instead. https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_CreateAssociation.html#systemsmanager-CreateAssociation-request-InstanceId",
			},
			"max_concurrency": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$`), "must be a valid number (e.g. 10) or percentage including the percent sign (e.g. 10%)"),
			},
			"max_errors": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$`), "must be a valid number (e.g. 10) or percentage including the percent sign (e.g. 10%)"),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"output_location": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrS3BucketName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(3, 63),
						},
						names.AttrS3KeyPrefix: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 500),
						},
						"s3_region": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(3, 20),
						},
					},
				},
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrScheduleExpression: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"sync_compliance": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AssociationSyncCompliance](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"targets": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 163),
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"wait_for_success_timeout_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ssm.CreateAssociationInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("apply_only_at_cron_interval"); ok {
		input.ApplyOnlyAtCronInterval = v.(bool)
	}

	if v, ok := d.GetOk("association_name"); ok {
		input.AssociationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("automation_target_parameter_name"); ok {
		input.AutomationTargetParameterName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("compliance_severity"); ok {
		input.ComplianceSeverity = awstypes.AssociationComplianceSeverity(v.(string))
	}

	if v, ok := d.GetOk("document_version"); ok {
		input.DocumentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrInstanceID); ok {
		input.InstanceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		input.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_errors"); ok {
		input.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("output_location"); ok {
		input.OutputLocation = expandAssociationOutputLocation(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrParameters); ok {
		input.Parameters = expandParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrScheduleExpression); ok {
		input.ScheduleExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sync_compliance"); ok {
		input.SyncCompliance = awstypes.AssociationSyncCompliance(v.(string))
	}

	if v, ok := d.GetOk("targets"); ok {
		input.Targets = expandTargets(v.([]interface{}))
	}

	output, err := conn.CreateAssociation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Association (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.AssociationDescription.AssociationId))

	if v, ok := d.GetOk("wait_for_success_timeout_seconds"); ok {
		timeout := time.Duration(v.(int)) * time.Second //nolint:durationcheck // should really be d.Timeout(schema.TimeoutCreate)
		if _, err := waitAssociationCreated(ctx, conn, d.Id(), timeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SSM Association (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAssociationRead(ctx, d, meta)...)
}

func resourceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	association, err := findAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Association (%s): %s", d.Id(), err)
	}

	d.Set("apply_only_at_cron_interval", association.ApplyOnlyAtCronInterval)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ssm",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "association/" + aws.ToString(association.AssociationId),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrAssociationID, association.AssociationId)
	d.Set("association_name", association.AssociationName)
	d.Set("automation_target_parameter_name", association.AutomationTargetParameterName)
	d.Set("compliance_severity", association.ComplianceSeverity)
	d.Set("document_version", association.DocumentVersion)
	d.Set(names.AttrInstanceID, association.InstanceId)
	d.Set("max_concurrency", association.MaxConcurrency)
	d.Set("max_errors", association.MaxErrors)
	d.Set(names.AttrName, association.Name)
	if err := d.Set("output_location", flattenAssociationOutputLocation(association.OutputLocation)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_location: %s", err)
	}
	if err := d.Set(names.AttrParameters, flattenParameters(association.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	d.Set(names.AttrScheduleExpression, association.ScheduleExpression)
	d.Set("sync_compliance", association.SyncCompliance)
	if err := d.Set("targets", flattenTargets(association.Targets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting targets: %s", err)
	}

	return diags
}

func resourceAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		// AWS creates a new version every time the association is updated, so everything should be passed in the update.
		input := &ssm.UpdateAssociationInput{
			AssociationId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("apply_only_at_cron_interval"); ok {
			input.ApplyOnlyAtCronInterval = v.(bool)
		}

		if v, ok := d.GetOk("association_name"); ok {
			input.AssociationName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("automation_target_parameter_name"); ok {
			input.AutomationTargetParameterName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("compliance_severity"); ok {
			input.ComplianceSeverity = awstypes.AssociationComplianceSeverity(v.(string))
		}

		if v, ok := d.GetOk("document_version"); ok {
			input.DocumentVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_concurrency"); ok {
			input.MaxConcurrency = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_errors"); ok {
			input.MaxErrors = aws.String(v.(string))
		}

		if v, ok := d.GetOk("output_location"); ok {
			input.OutputLocation = expandAssociationOutputLocation(v.([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok {
			input.Parameters = expandParameters(v.(map[string]interface{}))
		}

		if v, ok := d.GetOk(names.AttrScheduleExpression); ok {
			input.ScheduleExpression = aws.String(v.(string))
		}

		if d.HasChange("sync_compliance") {
			input.SyncCompliance = awstypes.AssociationSyncCompliance(d.Get("sync_compliance").(string))
		}

		if _, ok := d.GetOk("targets"); ok {
			input.Targets = expandTargets(d.Get("targets").([]interface{}))
		}

		_, err := conn.UpdateAssociation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSM Association (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAssociationRead(ctx, d, meta)...)
}

func resourceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	log.Printf("[DEBUG] Deleting SSM Association: %s", d.Id())
	_, err := conn.DeleteAssociation(ctx, &ssm.DeleteAssociationInput{
		AssociationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.AssociationDoesNotExist](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findAssociationByID(ctx context.Context, conn *ssm.Client, id string) (*awstypes.AssociationDescription, error) {
	input := &ssm.DescribeAssociationInput{
		AssociationId: aws.String(id),
	}

	output, err := conn.DescribeAssociation(ctx, input)

	if errs.IsA[*awstypes.AssociationDoesNotExist](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AssociationDescription == nil || output.AssociationDescription.Overview == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AssociationDescription, nil
}

func statusAssociation(ctx context.Context, conn *ssm.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// Use the Overview.Status field instead of the root-level Status as DescribeAssociation
		// does not appear to return the root-level Status in the API response at this time.
		return output, aws.ToString(output.Overview.Status), nil
	}
}

func waitAssociationCreated(ctx context.Context, conn *ssm.Client, id string, timeout time.Duration) (*awstypes.AssociationDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AssociationStatusNamePending),
		Target:  enum.Slice(awstypes.AssociationStatusNameSuccess),
		Refresh: statusAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AssociationDescription); ok {
		if status := awstypes.AssociationStatusName(aws.ToString(output.Overview.Status)); status == awstypes.AssociationStatusNameFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Overview.DetailedStatus)))
		}

		return output, err
	}

	return nil, err
}

func expandParameters(tfMap map[string]interface{}) map[string][]string {
	return tfmaps.ApplyToAllValues(tfMap, func(v interface{}) []string {
		return []string{v.(string)}
	})
}

func flattenParameters(apiObject map[string][]string) map[string]interface{} {
	return tfmaps.ApplyToAllValues(apiObject, func(v []string) interface{} {
		return strings.Join(v, ",")
	})
}

func expandAssociationOutputLocation(tfList []interface{}) *awstypes.InstanceAssociationOutputLocation {
	if tfList == nil {
		return nil
	}

	//We only allow 1 Item so we can grab the first in the list only
	tfMap := tfList[0].(map[string]interface{})

	s3OutputLocation := &awstypes.S3OutputLocation{
		OutputS3BucketName: aws.String(tfMap[names.AttrS3BucketName].(string)),
	}

	if v, ok := tfMap[names.AttrS3KeyPrefix]; ok {
		s3OutputLocation.OutputS3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := tfMap["s3_region"].(string); ok && v != "" {
		s3OutputLocation.OutputS3Region = aws.String(v)
	}

	return &awstypes.InstanceAssociationOutputLocation{
		S3Location: s3OutputLocation,
	}
}

func flattenAssociationOutputLocation(apiObject *awstypes.InstanceAssociationOutputLocation) []interface{} {
	if apiObject == nil || apiObject.S3Location == nil {
		return nil
	}

	tfList := make([]interface{}, 0)
	tfMap := make(map[string]interface{})

	tfMap[names.AttrS3BucketName] = aws.ToString(apiObject.S3Location.OutputS3BucketName)

	if apiObject.S3Location.OutputS3KeyPrefix != nil {
		tfMap[names.AttrS3KeyPrefix] = aws.ToString(apiObject.S3Location.OutputS3KeyPrefix)
	}

	if apiObject.S3Location.OutputS3Region != nil {
		tfMap["s3_region"] = aws.ToString(apiObject.S3Location.OutputS3Region)
	}

	tfList = append(tfList, tfMap)

	return tfList
}
