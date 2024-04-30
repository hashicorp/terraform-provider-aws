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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_backup_framework", name="Framework")
// @Tags(identifierAttribute="arn")
func ResourceFramework() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFrameworkCreate,
		ReadWithoutTimeout:   resourceFrameworkRead,
		UpdateWithoutTimeout: resourceFrameworkUpdate,
		DeleteWithoutTimeout: resourceFrameworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"arn": {
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
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"scope": {
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
										MinItems: 1,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"compliance_resource_types": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									// A maximum of one key-value pair can be provided.
									// The tag value is optional, but it cannot be an empty string
									"tags": tftags.TagsSchema(),
								},
							},
						},
					},
				},
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validFrameworkName,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFrameworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	name := d.Get("name").(string)
	input := &backup.CreateFrameworkInput{
		IdempotencyToken:  aws.String(id.UniqueId()),
		FrameworkControls: expandFrameworkControls(ctx, d.Get("control").(*schema.Set).List()),
		FrameworkName:     aws.String(name),
		FrameworkTags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.FrameworkDescription = aws.String(v.(string))
	}

	resp, err := conn.CreateFramework(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Framework: %s", err)
	}

	// Set ID with the name since the name is unique for the framework
	d.SetId(aws.ToString(resp.FrameworkName))

	// waiter since the status changes from CREATE_IN_PROGRESS to either COMPLETED or FAILED
	if _, err := waitFrameworkCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Framework (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceFrameworkRead(ctx, d, meta)...)
}

func resourceFrameworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	resp, err := conn.DescribeFramework(ctx, &backup.DescribeFrameworkInput{
		FrameworkName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Backup Framework (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Framework (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.FrameworkArn)
	d.Set("deployment_status", resp.DeploymentStatus)
	d.Set("description", resp.FrameworkDescription)
	d.Set("name", resp.FrameworkName)
	d.Set("status", resp.FrameworkStatus)

	if err := d.Set("creation_time", resp.CreationTime.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting creation_time: %s", err)
	}

	if err := d.Set("control", flattenFrameworkControls(ctx, resp.FrameworkControls)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting control: %s", err)
	}

	return diags
}

func resourceFrameworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	if d.HasChanges("description", "control") {
		input := &backup.UpdateFrameworkInput{
			IdempotencyToken:     aws.String(id.UniqueId()),
			FrameworkControls:    expandFrameworkControls(ctx, d.Get("control").(*schema.Set).List()),
			FrameworkDescription: aws.String(d.Get("description").(string)),
			FrameworkName:        aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating Backup Framework: %#v", input)

		_, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.UpdateFramework(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Backup Framework (%s): %s", d.Id(), err)
		}

		if _, err := waitFrameworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Framework (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFrameworkRead(ctx, d, meta)...)
}

func resourceFrameworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	input := &backup.DeleteFrameworkInput{
		FrameworkName: aws.String(d.Id()),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteFramework(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Framework (%s): %s", d.Id(), err)
	}

	if _, err := waitFrameworkDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Framework (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func expandFrameworkControls(ctx context.Context, controls []interface{}) []awstypes.FrameworkControl {
	if len(controls) == 0 {
		return nil
	}

	frameworkControls := []awstypes.FrameworkControl{}

	for _, control := range controls {
		tfMap := control.(map[string]interface{})

		// on some updates, there is an { ControlName: "" } element in Framework Controls.
		// this element must be skipped to avoid the "A control name is required." error
		// this happens for Step 7/7 for TestAccBackupFramework_updateControlScope
		if v, ok := tfMap["name"].(string); ok && v == "" {
			continue
		}

		frameworkControl := awstypes.FrameworkControl{
			ControlName:  aws.String(tfMap["name"].(string)),
			ControlScope: expandControlScope(ctx, tfMap["scope"].([]interface{})),
		}

		if v, ok := tfMap["input_parameter"]; ok && v.(*schema.Set).Len() > 0 {
			frameworkControl.ControlInputParameters = expandInputParameters(tfMap["input_parameter"].(*schema.Set).List())
		}

		frameworkControls = append(frameworkControls, frameworkControl)
	}

	return frameworkControls
}

func expandInputParameters(inputParams []interface{}) []awstypes.ControlInputParameter {
	if len(inputParams) == 0 {
		return nil
	}

	controlInputParameters := []awstypes.ControlInputParameter{}

	for _, inputParam := range inputParams {
		tfMap := inputParam.(map[string]interface{})
		controlInputParameter := awstypes.ControlInputParameter{}

		if v, ok := tfMap["name"].(string); ok && v != "" {
			controlInputParameter.ParameterName = aws.String(v)
		}

		if v, ok := tfMap["value"].(string); ok && v != "" {
			controlInputParameter.ParameterValue = aws.String(v)
		}

		controlInputParameters = append(controlInputParameters, controlInputParameter)
	}

	return controlInputParameters
}

func expandControlScope(ctx context.Context, scope []interface{}) *awstypes.ControlScope {
	if len(scope) == 0 || scope[0] == nil {
		return nil
	}

	tfMap, ok := scope[0].(map[string]interface{})
	if !ok {
		return nil
	}

	controlScope := &awstypes.ControlScope{}

	if v, ok := tfMap["compliance_resource_ids"]; ok && v.(*schema.Set).Len() > 0 {
		controlScope.ComplianceResourceIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["compliance_resource_types"]; ok && v.(*schema.Set).Len() > 0 {
		controlScope.ComplianceResourceTypes = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	// A maximum of one key-value pair can be provided.
	// The tag value is optional, but it cannot be an empty string
	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		controlScope.Tags = Tags(tftags.New(ctx, v).IgnoreAWS())
	}

	return controlScope
}

func flattenFrameworkControls(ctx context.Context, controls []awstypes.FrameworkControl) []interface{} {
	if controls == nil {
		return []interface{}{}
	}

	frameworkControls := []interface{}{}
	for _, control := range controls {
		values := map[string]interface{}{}
		values["input_parameter"] = flattenInputParameters(control.ControlInputParameters)
		values["name"] = aws.ToString(control.ControlName)
		values["scope"] = flattenScope(ctx, control.ControlScope)
		frameworkControls = append(frameworkControls, values)
	}
	return frameworkControls
}

func flattenInputParameters(inputParams []awstypes.ControlInputParameter) []interface{} {
	if inputParams == nil {
		return []interface{}{}
	}

	controlInputParameters := []interface{}{}
	for _, inputParam := range inputParams {
		values := map[string]interface{}{}
		values["name"] = aws.ToString(inputParam.ParameterName)
		values["value"] = aws.ToString(inputParam.ParameterValue)
		controlInputParameters = append(controlInputParameters, values)
	}
	return controlInputParameters
}

func flattenScope(ctx context.Context, scope *awstypes.ControlScope) []interface{} {
	if scope == nil {
		return []interface{}{}
	}

	controlScope := map[string]interface{}{
		"compliance_resource_ids":   flex.FlattenStringValueList(scope.ComplianceResourceIds),
		"compliance_resource_types": flex.FlattenStringValueList(scope.ComplianceResourceTypes),
	}

	if v := scope.Tags; v != nil {
		controlScope["tags"] = KeyValueTags(ctx, v).IgnoreAWS().Map()
	}

	return []interface{}{controlScope}
}
