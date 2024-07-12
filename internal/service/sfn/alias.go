// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sfn_alias", name="Alias")
func resourceAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAliasCreate,
		ReadWithoutTimeout:   resourceAliasRead,
		UpdateWithoutTimeout: resourceAliasUpdate,
		DeleteWithoutTimeout: resourceAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"routing_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"state_machine_version_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrWeight: {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

const (
	ResNameAlias = "Alias"
)

func resourceAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	in := &sfn.CreateStateMachineAliasInput{
		Name:        aws.String(d.Get(names.AttrName).(string)),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
	}

	if v, ok := d.GetOk("routing_configuration"); ok && len(v.([]interface{})) > 0 {
		in.RoutingConfiguration = expandAliasRoutingConfiguration(v.([]interface{}))
	}

	out, err := conn.CreateStateMachineAlias(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SFN, create.ErrActionCreating, ResNameAlias, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.StateMachineAliasArn == nil {
		return create.AppendDiagError(diags, names.SFN, create.ErrActionCreating, ResNameAlias, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.StateMachineAliasArn))

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	out, err := findAliasByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SFN Alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SFN, create.ErrActionReading, ResNameAlias, d.Id(), err)
	}

	d.Set(names.AttrARN, out.StateMachineAliasArn)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrCreationDate, aws.ToTime(out.CreationDate).Format(time.RFC3339))
	d.SetId(aws.ToString(out.StateMachineAliasArn))

	if err := d.Set("routing_configuration", flattenAliasRoutingConfiguration(out.RoutingConfiguration)); err != nil {
		return create.AppendDiagError(diags, names.SFN, create.ErrActionSetting, ResNameAlias, d.Id(), err)
	}
	return diags
}

func resourceAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	update := false

	in := &sfn.UpdateStateMachineAliasInput{
		StateMachineAliasArn: aws.String(d.Id()),
	}

	if d.HasChanges(names.AttrDescription) {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
		update = true
	}

	if d.HasChange("routing_configuration") {
		in.RoutingConfiguration = expandAliasRoutingConfiguration(d.Get("routing_configuration").([]interface{}))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating SFN Alias (%s): %#v", d.Id(), in)
	_, err := conn.UpdateStateMachineAlias(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SFN, create.ErrActionUpdating, ResNameAlias, d.Id(), err)
	}

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	log.Printf("[INFO] Deleting SFN Alias %s", d.Id())
	_, err := conn.DeleteStateMachineAlias(ctx, &sfn.DeleteStateMachineAliasInput{
		StateMachineAliasArn: aws.String(d.Id()),
	})

	if err != nil {
		return create.AppendDiagError(diags, names.SFN, create.ErrActionDeleting, ResNameAlias, d.Id(), err)
	}

	return diags
}

func findAliasByARN(ctx context.Context, conn *sfn.Client, arn string) (*sfn.DescribeStateMachineAliasOutput, error) {
	in := &sfn.DescribeStateMachineAliasInput{
		StateMachineAliasArn: aws.String(arn),
	}
	out, err := conn.DescribeStateMachineAlias(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenAliasRoutingConfigurationItem(apiObject awstypes.RoutingConfigurationListItem) map[string]interface{} {
	tfMap := map[string]interface{}{
		names.AttrWeight: apiObject.Weight,
	}

	if v := apiObject.StateMachineVersionArn; v != nil {
		tfMap["state_machine_version_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenAliasRoutingConfiguration(apiObjects []awstypes.RoutingConfigurationListItem) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenAliasRoutingConfigurationItem(apiObject))
	}

	return tfList
}

func expandAliasRoutingConfiguration(tfList []interface{}) []awstypes.RoutingConfigurationListItem {
	if len(tfList) == 0 {
		return nil
	}
	var configurationListItems []awstypes.RoutingConfigurationListItem

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		configurationListItem := expandAliasRoutingConfigurationItem(tfMap)

		if configurationListItem.StateMachineVersionArn == nil {
			continue
		}

		configurationListItems = append(configurationListItems, configurationListItem)
	}

	return configurationListItems
}

func expandAliasRoutingConfigurationItem(tfMap map[string]interface{}) awstypes.RoutingConfigurationListItem {
	apiObject := awstypes.RoutingConfigurationListItem{}

	if v, ok := tfMap["state_machine_version_arn"].(string); ok && v != "" {
		apiObject.StateMachineVersionArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrWeight].(int); ok && v != 0 {
		apiObject.Weight = int32(v)
	}

	return apiObject
}
