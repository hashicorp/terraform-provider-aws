// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/aws/aws-sdk-go-v2/service/codecommit/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codecommit_trigger", name="Trigger")
func resourceTrigger() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTriggerCreate,
		ReadWithoutTimeout:   resourceTriggerRead,
		DeleteWithoutTimeout: resourceTriggerDelete,

		Schema: map[string]*schema.Schema{
			"configuration_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRepositoryName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"trigger": {
				Type:     schema.TypeSet,
				ForceNew: true,
				Required: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branches": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"custom_data": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrDestinationARN: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"events": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.RepositoryTriggerEventEnum](),
							},
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceTriggerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	repositoryName := d.Get(names.AttrRepositoryName).(string)
	input := &codecommit.PutRepositoryTriggersInput{
		RepositoryName: aws.String(repositoryName),
		Triggers:       expandRepositoryTriggers(d.Get("trigger").(*schema.Set).List()),
	}

	_, err := conn.PutRepositoryTriggers(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeCommit Trigger (%s): %s", repositoryName, err)
	}

	d.SetId(repositoryName)

	return append(diags, resourceTriggerRead(ctx, d, meta)...)
}

func resourceTriggerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	output, err := findRepositoryTriggersByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeCommit Trigger %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Trigger (%s): %s", d.Id(), err)
	}

	d.Set("configuration_id", output.ConfigurationId)
	d.Set(names.AttrRepositoryName, d.Id())
	if err := d.Set("trigger", flattenRepositoryTriggers(output.Triggers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trigger: %s", err)
	}

	return diags
}

func resourceTriggerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	log.Printf("[DEBUG] Deleting CodeCommit Trigger: %s", d.Id())
	input := &codecommit.PutRepositoryTriggersInput{
		RepositoryName: aws.String(d.Id()),
		Triggers:       []types.RepositoryTrigger{},
	}

	_, err := conn.PutRepositoryTriggers(ctx, input)

	if errs.IsA[*types.RepositoryDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeCommit Trigger (%s): %s", d.Id(), err)
	}

	return diags
}

func findRepositoryTriggersByName(ctx context.Context, conn *codecommit.Client, repositoryName string) (*codecommit.GetRepositoryTriggersOutput, error) {
	input := &codecommit.GetRepositoryTriggersInput{
		RepositoryName: aws.String(repositoryName),
	}

	output, err := conn.GetRepositoryTriggers(ctx, input)

	if errs.IsA[*types.RepositoryDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Triggers) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandRepositoryTriggers(tfList []interface{}) []types.RepositoryTrigger {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]types.RepositoryTrigger, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.RepositoryTrigger{}

		// "RepositoryTriggerBranchNameListRequiredException: Repository trigger branch name list cannot be null".
		if v, ok := tfMap["branches"].([]interface{}); ok {
			apiObject.Branches = flex.ExpandStringValueList(v)
		}

		if v, ok := tfMap["custom_data"].(string); ok && v != "" {
			apiObject.CustomData = aws.String(v)
		}

		if v, ok := tfMap[names.AttrDestinationARN].(string); ok && v != "" {
			apiObject.DestinationArn = aws.String(v)
		}

		if v, ok := tfMap["events"].([]interface{}); ok && len(v) > 0 {
			apiObject.Events = flex.ExpandStringyValueList[types.RepositoryTriggerEventEnum](v)
		}

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenRepositoryTriggers(apiObjects []types.RepositoryTrigger) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if v := apiObject.Branches; v != nil {
			tfMap["branches"] = v
		}

		if v := apiObject.CustomData; v != nil {
			tfMap["custom_data"] = aws.ToString(v)
		}

		if v := apiObject.DestinationArn; v != nil {
			tfMap[names.AttrDestinationARN] = aws.ToString(v)
		}

		if v := apiObject.Events; v != nil {
			tfMap["events"] = v
		}

		if v := apiObject.Name; v != nil {
			tfMap[names.AttrName] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
