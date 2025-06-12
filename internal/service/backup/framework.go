// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_backup_framework", name="Framework")
// @Tags(identifierAttribute="arn")
// @Testing(serialize=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/backup;backup.DescribeFrameworkOutput")
// @Testing(generator="randomFrameworkName()")
func resourceFramework() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFrameworkCreate,
		ReadWithoutTimeout:   resourceFrameworkRead,
		UpdateWithoutTimeout: resourceFrameworkUpdate,
		DeleteWithoutTimeout: resourceFrameworkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"control": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_parameter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						names.AttrScope: {
							// The control scope can include
							// one or more resource types,
							// a combination of a tag key and value,
							// or a combination of one resource type and one resource ID.
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compliance_resource_ids": {
										Type:     schema.TypeSet,
										Optional: true,
										Computed: true,
										MinItems: 1,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"compliance_resource_types": {
										Type:     schema.TypeSet,
										Optional: true,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									// A maximum of one key-value pair can be provided.
									// The tag value is optional, but it cannot be an empty string
									names.AttrTags: tftags.TagsSchema(),
								},
							},
						},
					},
				},
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
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
				ForceNew:     true,
				ValidateFunc: validFrameworkName,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFrameworkCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &backup.CreateFrameworkInput{
		FrameworkControls: expandFrameworkControls(ctx, d.Get("control").(*schema.Set).List()),
		FrameworkName:     aws.String(name),
		FrameworkTags:     getTagsIn(ctx),
		IdempotencyToken:  aws.String(sdkid.UniqueId()),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.FrameworkDescription = aws.String(v.(string))
	}

	_, err := conn.CreateFramework(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Framework (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitFrameworkCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for  Backup Framework (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFrameworkRead(ctx, d, meta)...)
}

func resourceFrameworkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	output, err := findFrameworkByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Framework (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Framework (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.FrameworkArn)
	if err := d.Set("control", flattenFrameworkControls(ctx, output.FrameworkControls)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting control: %s", err)
	}
	d.Set(names.AttrCreationTime, output.CreationTime.Format(time.RFC3339))
	d.Set("deployment_status", output.DeploymentStatus)
	d.Set(names.AttrDescription, output.FrameworkDescription)
	d.Set(names.AttrName, output.FrameworkName)
	d.Set(names.AttrStatus, output.FrameworkStatus)

	return diags
}

func resourceFrameworkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	if d.HasChanges("control", names.AttrDescription) {
		input := &backup.UpdateFrameworkInput{
			FrameworkControls:    expandFrameworkControls(ctx, d.Get("control").(*schema.Set).List()),
			FrameworkDescription: aws.String(d.Get(names.AttrDescription).(string)),
			FrameworkName:        aws.String(d.Id()),
			IdempotencyToken:     aws.String(sdkid.UniqueId()),
		}

		_, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
			return conn.UpdateFramework(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Backup Framework (%s): %s", d.Id(), err)
		}

		if _, err := waitFrameworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Backup Framework (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFrameworkRead(ctx, d, meta)...)
}

func resourceFrameworkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	log.Printf("[DEBUG] Deleting Backup Framework: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, d.Timeout(schema.TimeoutDelete), func() (any, error) {
		return conn.DeleteFramework(ctx, &backup.DeleteFrameworkInput{
			FrameworkName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Framework (%s): %s", d.Id(), err)
	}

	if _, err := waitFrameworkDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Backup Framework (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFrameworkByName(ctx context.Context, conn *backup.Client, name string) (*backup.DescribeFrameworkOutput, error) {
	input := &backup.DescribeFrameworkInput{
		FrameworkName: aws.String(name),
	}

	return findFramework(ctx, conn, input)
}

func findFramework(ctx context.Context, conn *backup.Client, input *backup.DescribeFrameworkInput) (*backup.DescribeFrameworkOutput, error) {
	output, err := conn.DescribeFramework(ctx, input)

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

func statusFramework(ctx context.Context, conn *backup.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFrameworkByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.DeploymentStatus), nil
	}
}

const (
	frameworkStatusCompleted          = "COMPLETED"
	frameworkStatusCreationInProgress = "CREATE_IN_PROGRESS"
	frameworkStatusDeletionInProgress = "DELETE_IN_PROGRESS"
	frameworkStatusFailed             = "FAILED"
	frameworkStatusUpdateInProgress   = "UPDATE_IN_PROGRESS"
)

func waitFrameworkCreated(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*backup.DescribeFrameworkOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{frameworkStatusCreationInProgress},
		Target:  []string{frameworkStatusCompleted, frameworkStatusFailed},
		Refresh: statusFramework(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeFrameworkOutput); ok {
		return output, err
	}

	return nil, err
}

func waitFrameworkUpdated(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*backup.DescribeFrameworkOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{frameworkStatusUpdateInProgress},
		Target:  []string{frameworkStatusCompleted, frameworkStatusFailed},
		Refresh: statusFramework(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeFrameworkOutput); ok {
		return output, err
	}

	return nil, err
}

func waitFrameworkDeleted(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*backup.DescribeFrameworkOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{frameworkStatusDeletionInProgress},
		Target:  []string{},
		Refresh: statusFramework(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeFrameworkOutput); ok {
		return output, err
	}

	return nil, err
}

func expandFrameworkControls(ctx context.Context, tfList []any) []awstypes.FrameworkControl {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []awstypes.FrameworkControl{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		// on some updates, there is an { ControlName: "" } element in Framework Controls.
		// this element must be skipped to avoid the "A control name is required." error
		// this happens for Step 7/7 for TestAccBackupFramework_updateControlScope
		if v, ok := tfMap[names.AttrName].(string); ok && v == "" {
			continue
		}

		apiObject := awstypes.FrameworkControl{
			ControlName:  aws.String(tfMap[names.AttrName].(string)),
			ControlScope: expandControlScope(ctx, tfMap[names.AttrScope].([]any)),
		}

		if v, ok := tfMap["input_parameter"]; ok && v.(*schema.Set).Len() > 0 {
			apiObject.ControlInputParameters = expandControlInputParameters(tfMap["input_parameter"].(*schema.Set).List())
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandControlInputParameters(tfList []any) []awstypes.ControlInputParameter {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []awstypes.ControlInputParameter{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.ControlInputParameter{}

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.ParameterName = aws.String(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.ParameterValue = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandControlScope(ctx context.Context, tfList []any) *awstypes.ControlScope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ControlScope{}

	if v, ok := tfMap["compliance_resource_ids"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.ComplianceResourceIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["compliance_resource_types"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.ComplianceResourceTypes = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	// A maximum of one key-value pair can be provided.
	// The tag value is optional, but it cannot be an empty string
	if v, ok := tfMap[names.AttrTags].(map[string]any); ok && len(v) > 0 {
		apiObject.Tags = svcTags(tftags.New(ctx, v).IgnoreAWS())
	}

	return apiObject
}

func flattenFrameworkControls(ctx context.Context, apiObjects []awstypes.FrameworkControl) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}
		tfMap["input_parameter"] = flattenControlInputParameters(apiObject.ControlInputParameters)
		tfMap[names.AttrName] = aws.ToString(apiObject.ControlName)
		tfMap[names.AttrScope] = flattenControlScope(ctx, apiObject.ControlScope)

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenControlInputParameters(apiObjects []awstypes.ControlInputParameter) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}
		tfMap[names.AttrName] = aws.ToString(apiObject.ParameterName)
		tfMap[names.AttrValue] = aws.ToString(apiObject.ParameterValue)

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenControlScope(ctx context.Context, apiObject *awstypes.ControlScope) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"compliance_resource_ids":   apiObject.ComplianceResourceIds,
		"compliance_resource_types": apiObject.ComplianceResourceTypes,
	}

	if v := apiObject.Tags; v != nil {
		tfMap[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
	}

	return []any{tfMap}
}
