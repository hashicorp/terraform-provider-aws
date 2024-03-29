// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sfn_alias")
func ResourceAlias() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
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
						"weight": {
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
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	in := &sfn.CreateStateMachineAliasInput{
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
	}

	if v, ok := d.GetOk("routing_configuration"); ok && len(v.([]interface{})) > 0 {
		in.RoutingConfiguration = expandAliasRoutingConfiguration(v.([]interface{}))
	}

	out, err := conn.CreateStateMachineAliasWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.SFN, create.ErrActionCreating, ResNameAlias, d.Get("name").(string), err)
	}

	if out == nil || out.StateMachineAliasArn == nil {
		return create.DiagError(names.SFN, create.ErrActionCreating, ResNameAlias, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.StateMachineAliasArn))

	return resourceAliasRead(ctx, d, meta)
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	out, err := FindAliasByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SFN Alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SFN, create.ErrActionReading, ResNameAlias, d.Id(), err)
	}

	d.Set("arn", out.StateMachineAliasArn)
	d.Set("name", out.Name)
	d.Set("description", out.Description)
	d.Set("creation_date", aws.TimeValue(out.CreationDate).Format(time.RFC3339))
	d.SetId(aws.StringValue(out.StateMachineAliasArn))

	if err := d.Set("routing_configuration", flattenAliasRoutingConfiguration(out.RoutingConfiguration)); err != nil {
		return create.DiagError(names.SFN, create.ErrActionSetting, ResNameAlias, d.Id(), err)
	}
	return nil
}

func resourceAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	update := false

	in := &sfn.UpdateStateMachineAliasInput{
		StateMachineAliasArn: aws.String(d.Id()),
	}

	if d.HasChanges("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChange("routing_configuration") {
		in.RoutingConfiguration = expandAliasRoutingConfiguration(d.Get("routing_configuration").([]interface{}))
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating SFN Alias (%s): %#v", d.Id(), in)
	_, err := conn.UpdateStateMachineAliasWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.SFN, create.ErrActionUpdating, ResNameAlias, d.Id(), err)
	}

	return resourceAliasRead(ctx, d, meta)
}

func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)
	log.Printf("[INFO] Deleting SFN Alias %s", d.Id())

	_, err := conn.DeleteStateMachineAliasWithContext(ctx, &sfn.DeleteStateMachineAliasInput{
		StateMachineAliasArn: aws.String(d.Id()),
	})

	if err != nil {
		return create.DiagError(names.SFN, create.ErrActionDeleting, ResNameAlias, d.Id(), err)
	}

	return nil
}

func FindAliasByARN(ctx context.Context, conn *sfn.SFN, arn string) (*sfn.DescribeStateMachineAliasOutput, error) {
	in := &sfn.DescribeStateMachineAliasInput{
		StateMachineAliasArn: aws.String(arn),
	}
	out, err := conn.DescribeStateMachineAliasWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, sfn.ErrCodeResourceNotFound) {
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

func flattenAliasRoutingConfigurationItem(apiObject *sfn.RoutingConfigurationListItem) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.StateMachineVersionArn; v != nil {
		tfMap["state_machine_version_arn"] = aws.StringValue(v)
	}

	if v := apiObject.Weight; v != nil {
		tfMap["weight"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenAliasRoutingConfiguration(apiObjects []*sfn.RoutingConfigurationListItem) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenAliasRoutingConfigurationItem(apiObject))
	}

	return tfList
}

func expandAliasRoutingConfiguration(tfList []interface{}) []*sfn.RoutingConfigurationListItem {
	if len(tfList) == 0 {
		return nil
	}
	var configurationListItems []*sfn.RoutingConfigurationListItem

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		configurationListItem := expandAliasRoutingConfigurationItem(tfMap)

		if configurationListItem == nil {
			continue
		}

		configurationListItems = append(configurationListItems, configurationListItem)
	}

	return configurationListItems
}

func expandAliasRoutingConfigurationItem(tfMap map[string]interface{}) *sfn.RoutingConfigurationListItem {
	if tfMap == nil {
		return nil
	}

	apiObject := &sfn.RoutingConfigurationListItem{}
	if v, ok := tfMap["state_machine_version_arn"].(string); ok && v != "" {
		apiObject.StateMachineVersionArn = aws.String(v)
	}

	if v, ok := tfMap["weight"].(int); ok && v != 0 {
		apiObject.Weight = aws.Int64(int64(v))
	}

	return apiObject
}
