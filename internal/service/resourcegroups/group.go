// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcegroups

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroups"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroups/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_resourcegroups_group", name="Group")
// @Tags(identifierAttribute="arn")
func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConfiguration: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrParameters: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrValues: {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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
			"resource_query": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      types.QueryTypeTagFilters10,
							ValidateFunc: validation.StringInSlice(enum.Slice(types.QueryTypeTagFilters10), false),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ResourceGroupsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &resourcegroups.CreateGroupInput{
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	waitForConfigurationAttached := false
	if groupCfg, set := d.GetOk(names.AttrConfiguration); set {
		// Only expand and add configuration if its set
		input.Configuration = expandGroupConfigurationItems(groupCfg.(*schema.Set).List())
		waitForConfigurationAttached = true
	}

	if resourceQuery, set := d.GetOk("resource_query"); set {
		// Only expand and add resource query if its set
		input.ResourceQuery = expandResourceQuery(resourceQuery.([]interface{}))
	}

	output, err := conn.CreateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Resource Groups Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Group.Name))

	if waitForConfigurationAttached {
		if _, err := waitGroupConfigurationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Resource Groups Group (%s) configuration update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ResourceGroupsClient(ctx)

	group, err := findGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Resource Groups Group %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Resource Groups Group (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(group.GroupArn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, group.Description)
	d.Set(names.AttrName, group.Name)

	q, err := conn.GetGroupQuery(ctx, &resourcegroups.GetGroupQueryInput{
		GroupName: aws.String(d.Id()),
	})

	hasQuery := true
	if err != nil {
		if errs.IsA[*types.BadRequestException](err) {
			// Attempting to get the query on a configuration group returns BadRequestException.
			hasQuery = false
		} else {
			return sdkdiag.AppendErrorf(diags, "reading Resource Groups Group (%s) resource query: %s", d.Id(), err)
		}
	}

	groupCfg, err := findGroupConfigurationByGroupName(ctx, conn, d.Id())

	hasConfiguration := true
	if err != nil {
		if errs.IsA[*types.BadRequestException](err) {
			// Attempting to get configuration on a query group returns BadRequestException.
			hasConfiguration = false
		} else {
			return sdkdiag.AppendErrorf(diags, "reading Resource Groups Group (%s) configuration: %s", d.Id(), err)
		}
	}

	if hasQuery {
		resultQuery := map[string]interface{}{}
		resultQuery["query"] = aws.ToString(q.GroupQuery.ResourceQuery.Query)
		resultQuery[names.AttrType] = q.GroupQuery.ResourceQuery.Type
		if err := d.Set("resource_query", []map[string]interface{}{resultQuery}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting resource_query: %s", err)
		}
	}
	if hasConfiguration {
		if err := d.Set(names.AttrConfiguration, flattenGroupConfigurationItems(groupCfg.Configuration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
		}
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ResourceGroupsClient(ctx)

	// Conversion between a resource-query and configuration group is not possible and vice-versa
	if d.HasChange(names.AttrConfiguration) && d.HasChange("resource_query") {
		return sdkdiag.AppendErrorf(diags, "conversion between resource-query and configuration group types is not possible")
	}

	if d.HasChange(names.AttrDescription) {
		input := &resourcegroups.UpdateGroupInput{
			Description: aws.String(d.Get(names.AttrDescription).(string)),
			GroupName:   aws.String(d.Id()),
		}

		_, err := conn.UpdateGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Resource Groups Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("resource_query") {
		input := &resourcegroups.UpdateGroupQueryInput{
			GroupName:     aws.String(d.Id()),
			ResourceQuery: expandResourceQuery(d.Get("resource_query").([]interface{})),
		}

		_, err := conn.UpdateGroupQuery(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Resource Groups Group (%s) resource query: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrConfiguration) {
		input := &resourcegroups.PutGroupConfigurationInput{
			Configuration: expandGroupConfigurationItems(d.Get(names.AttrConfiguration).(*schema.Set).List()),
			Group:         aws.String(d.Id()),
		}

		_, err := conn.PutGroupConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Resource Groups Group (%s) configuration: %s", d.Id(), err)
		}

		if _, err := waitGroupConfigurationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Resource Groups Group (%s) configuration update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ResourceGroupsClient(ctx)

	log.Printf("[DEBUG] Deleting Resource Groups Group: %s", d.Id())
	_, err := conn.DeleteGroup(ctx, &resourcegroups.DeleteGroupInput{
		GroupName: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Resource Groups Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findGroupByName(ctx context.Context, conn *resourcegroups.Client, name string) (*types.Group, error) {
	input := &resourcegroups.GetGroupInput{
		GroupName: aws.String(name),
	}

	output, err := conn.GetGroup(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Group == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Group, nil
}

func findGroupConfigurationByGroupName(ctx context.Context, conn *resourcegroups.Client, groupName string) (*types.GroupConfiguration, error) {
	input := &resourcegroups.GetGroupConfigurationInput{
		Group: aws.String(groupName),
	}

	output, err := conn.GetGroupConfiguration(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.GroupConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GroupConfiguration, nil
}

func statusGroupConfiguration(ctx context.Context, conn *resourcegroups.Client, groupName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findGroupConfigurationByGroupName(ctx, conn, groupName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitGroupConfigurationUpdated(ctx context.Context, conn *resourcegroups.Client, groupName string, timeout time.Duration) (*types.GroupConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.GroupConfigurationStatusUpdating),
		Target:  enum.Slice(types.GroupConfigurationStatusUpdateComplete),
		Refresh: statusGroupConfiguration(ctx, conn, groupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.GroupConfiguration); ok {
		if status := output.Status; status == types.GroupConfigurationStatusUpdateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func expandGroupConfigurationParameters(parameterList []interface{}) []types.GroupConfigurationParameter {
	var parameters []types.GroupConfigurationParameter

	for _, param := range parameterList {
		parameter := param.(map[string]interface{})
		var values []string
		for _, val := range parameter[names.AttrValues].([]interface{}) {
			values = append(values, val.(string))
		}
		parameters = append(parameters, types.GroupConfigurationParameter{
			Name:   aws.String(parameter[names.AttrName].(string)),
			Values: values,
		})
	}

	return parameters
}

func expandGroupConfigurationItems(configurationItemList []interface{}) []types.GroupConfigurationItem {
	var configurationItems []types.GroupConfigurationItem

	for _, configItem := range configurationItemList {
		configItemMap := configItem.(map[string]interface{})
		configurationItems = append(configurationItems, types.GroupConfigurationItem{
			Parameters: expandGroupConfigurationParameters(configItemMap[names.AttrParameters].(*schema.Set).List()),
			Type:       aws.String(configItemMap[names.AttrType].(string)),
		})
	}

	return configurationItems
}

func expandResourceQuery(resourceQueryList []interface{}) *types.ResourceQuery {
	resourceQuery := resourceQueryList[0].(map[string]interface{})

	return &types.ResourceQuery{
		Query: aws.String(resourceQuery["query"].(string)),
		Type:  types.QueryType(resourceQuery[names.AttrType].(string)),
	}
}

func flattenGroupConfigurationParameter(param types.GroupConfigurationParameter) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := param.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(param.Name)
	}

	if v := param.Values; v != nil {
		tfMap[names.AttrValues] = v
	}

	return tfMap
}

func flattenGroupConfigurationItem(configuration types.GroupConfigurationItem) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := configuration.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	if v := configuration.Parameters; v != nil {
		var params []interface{}
		for _, param := range v {
			params = append(params, flattenGroupConfigurationParameter(param))
		}
		tfMap[names.AttrParameters] = params
	}

	return tfMap
}

func flattenGroupConfigurationItems(configurationItems []types.GroupConfigurationItem) []interface{} {
	if len(configurationItems) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, configuration := range configurationItems {
		tfList = append(tfList, flattenGroupConfigurationItem(configuration))
	}

	return tfList
}
