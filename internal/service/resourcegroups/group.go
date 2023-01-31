package resourcegroups

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGroup() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"resource_query"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameters": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"values": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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
			"resource_query": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				ConflictsWith: []string{"configuration"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  resourcegroups.QueryTypeTagFilters10,
							ValidateFunc: validation.StringInSlice([]string{
								resourcegroups.QueryTypeTagFilters10,
							}, false),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &resourcegroups.CreateGroupInput{
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(name),
		Tags:        Tags(tags.IgnoreAWS()),
	}

	waitForConfigurationAttached := false
	if groupCfg, set := d.GetOk("configuration"); set {
		// Only expand and add configuration if its set
		input.Configuration = extractResourceGroupConfigurationItems(groupCfg.(*schema.Set).List())
		waitForConfigurationAttached = true
	}

	if resourceQuery, set := d.GetOk("resource_query"); set {
		// Only expand and add resource query if its set
		input.ResourceQuery = extractResourceGroupResourceQuery(resourceQuery.([]interface{}))
	}

	output, err := conn.CreateGroupWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Resource Groups Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Group.Name))

	if waitForConfigurationAttached {
		if _, err := waitGroupConfigurationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("waiting for Resource Groups Group (%s) configuration update: %s", d.Id(), err)
		}
	}

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	group, err := FindGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Resource Groups Group %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Resource Groups Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(group.GroupArn)
	d.Set("arn", arn)
	d.Set("description", group.Description)
	d.Set("name", group.Name)

	q, err := conn.GetGroupQueryWithContext(ctx, &resourcegroups.GetGroupQueryInput{
		GroupName: aws.String(d.Id()),
	})

	isConfigurationGroup := false
	if err != nil {
		if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeBadRequestException) {
			// Attempting to get the query on a configuration group returns BadRequestException.
			isConfigurationGroup = true
		} else {
			return diag.Errorf("reading Resource Groups Group (%s) resource query: %s", d.Id(), err)
		}
	}

	if !isConfigurationGroup {
		resultQuery := map[string]interface{}{}
		resultQuery["query"] = aws.StringValue(q.GroupQuery.ResourceQuery.Query)
		resultQuery["type"] = aws.StringValue(q.GroupQuery.ResourceQuery.Type)
		if err := d.Set("resource_query", []map[string]interface{}{resultQuery}); err != nil {
			return diag.Errorf("setting resource_query: %s", err)
		}
	}

	if isConfigurationGroup {
		groupCfg, err := findGroupConfigurationByGroupName(ctx, conn, d.Id())

		if err != nil {
			return diag.Errorf("reading Resource Groups Group (%s) configuration: %s", d.Id(), err)
		}

		if err := d.Set("configuration", flattenResourceGroupConfigurationItems(groupCfg.Configuration)); err != nil {
			return diag.Errorf("setting configuration: %s", err)
		}
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Resource Groups Group (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn()

	// Conversion between a resource-query and configuration group is not possible and vice-versa
	if d.HasChange("configuration") && d.HasChange("resource_query") {
		return diag.Errorf("conversion between resource-query and configuration group types is not possible")
	}

	if d.HasChange("description") {
		input := &resourcegroups.UpdateGroupInput{
			Description: aws.String(d.Get("description").(string)),
			GroupName:   aws.String(d.Id()),
		}

		_, err := conn.UpdateGroupWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Resource Groups Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("resource_query") {
		input := &resourcegroups.UpdateGroupQueryInput{
			GroupName:     aws.String(d.Id()),
			ResourceQuery: extractResourceGroupResourceQuery(d.Get("resource_query").([]interface{})),
		}

		_, err := conn.UpdateGroupQueryWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Resource Groups Group (%s) resource query: %s", d.Id(), err)
		}
	}

	if d.HasChange("configuration") {
		input := &resourcegroups.PutGroupConfigurationInput{
			Configuration: extractResourceGroupConfigurationItems(d.Get("configuration").(*schema.Set).List()),
			Group:         aws.String(d.Id()),
		}

		_, err := conn.PutGroupConfigurationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Resource Groups Group (%s) configuration: %s", d.Id(), err)
		}

		if _, err := waitGroupConfigurationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Resource Groups Group (%s) configuration update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating tags: %s", err)
		}
	}

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn()

	log.Printf("[DEBUG] Deleting Resource Groups Group: %s", d.Id())
	_, err := conn.DeleteGroupWithContext(ctx, &resourcegroups.DeleteGroupInput{
		GroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Resource Groups Group (%s): %s", d.Id(), err)
	}

	return nil
}

func FindGroupByName(ctx context.Context, conn *resourcegroups.ResourceGroups, name string) (*resourcegroups.Group, error) {
	input := &resourcegroups.GetGroupInput{
		GroupName: aws.String(name),
	}

	output, err := conn.GetGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
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

func findGroupConfigurationByGroupName(ctx context.Context, conn *resourcegroups.ResourceGroups, groupName string) (*resourcegroups.GroupConfiguration, error) {
	input := &resourcegroups.GetGroupConfigurationInput{
		Group: aws.String(groupName),
	}

	output, err := conn.GetGroupConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
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

func statusGroupConfiguration(ctx context.Context, conn *resourcegroups.ResourceGroups, groupName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findGroupConfigurationByGroupName(ctx, conn, groupName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitGroupConfigurationUpdated(ctx context.Context, conn *resourcegroups.ResourceGroups, groupName string, timeout time.Duration) (*resourcegroups.GroupConfiguration, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{resourcegroups.GroupConfigurationStatusUpdating},
		Target:  []string{resourcegroups.GroupConfigurationStatusUpdateComplete},
		Refresh: statusGroupConfiguration(ctx, conn, groupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*resourcegroups.GroupConfiguration); ok {
		if status := aws.StringValue(output.Status); status == resourcegroups.GroupConfigurationStatusUpdateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func extractResourceGroupConfigurationParameters(parameterList []interface{}) []*resourcegroups.GroupConfigurationParameter {
	var parameters []*resourcegroups.GroupConfigurationParameter

	for _, param := range parameterList {
		parameter := param.(map[string]interface{})
		var values []string
		for _, val := range parameter["values"].([]interface{}) {
			values = append(values, val.(string))
		}
		parameters = append(parameters, &resourcegroups.GroupConfigurationParameter{
			Name:   aws.String(parameter["name"].(string)),
			Values: aws.StringSlice(values),
		})
	}

	return parameters
}

func extractResourceGroupConfigurationItems(configurationItemList []interface{}) []*resourcegroups.GroupConfigurationItem {
	var configurationItems []*resourcegroups.GroupConfigurationItem

	for _, configItem := range configurationItemList {
		configItemMap := configItem.(map[string]interface{})
		configurationItems = append(configurationItems, &resourcegroups.GroupConfigurationItem{
			Parameters: extractResourceGroupConfigurationParameters(configItemMap["parameters"].(*schema.Set).List()),
			Type:       aws.String(configItemMap["type"].(string)),
		})
	}

	return configurationItems
}

func flattenResourceGroupConfigurationParameter(param *resourcegroups.GroupConfigurationParameter) map[string]interface{} {
	if param == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := param.Name; v != nil {
		tfMap["name"] = aws.StringValue(param.Name)
	}

	if v := param.Values; v != nil {
		tfMap["values"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenResourceGroupConfigurationItem(configuration *resourcegroups.GroupConfigurationItem) map[string]interface{} {
	if configuration == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := configuration.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := configuration.Parameters; v != nil {
		var params []interface{}
		for _, param := range v {
			params = append(params, flattenResourceGroupConfigurationParameter(param))
		}
		tfMap["parameters"] = params
	}

	return tfMap
}

func flattenResourceGroupConfigurationItems(configurationItems []*resourcegroups.GroupConfigurationItem) []interface{} {
	if len(configurationItems) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, configuration := range configurationItems {
		if configuration == nil {
			continue
		}

		tfList = append(tfList, flattenResourceGroupConfigurationItem(configuration))
	}

	return tfList
}

func extractResourceGroupResourceQuery(resourceQueryList []interface{}) *resourcegroups.ResourceQuery {
	resourceQuery := resourceQueryList[0].(map[string]interface{})

	return &resourcegroups.ResourceQuery{
		Query: aws.String(resourceQuery["query"].(string)),
		Type:  aws.String(resourceQuery["type"].(string)),
	}
}
