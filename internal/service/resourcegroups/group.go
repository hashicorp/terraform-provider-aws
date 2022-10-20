package resourcegroups

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for a group configuration to be attached
	groupConfigurationAttachedTimeout = 15 * time.Minute
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
							Required: true,
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

func resourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := resourcegroups.CreateGroupInput{
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(d.Get("name").(string)),
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

	res, err := conn.CreateGroup(&input)
	if err != nil {
		return fmt.Errorf("error creating resource group: %s", err)
	}

	if waitForConfigurationAttached {
		// Need to wait and refresh for when the configuration has been attached to the group
		err := waitForConfigurationUpdatedState(conn, aws.StringValue(res.Group.Name), groupConfigurationAttachedTimeout)
		if err != nil {
			return fmt.Errorf("error attaching configuration to resource group: %s", err)
		}
	}

	d.SetId(aws.StringValue(res.Group.Name))

	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	group, err := FindGroupByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Resource Groups Group %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Resource Groups Group (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(group.GroupArn)
	d.Set("arn", arn)
	d.Set("description", group.Description)
	d.Set("name", group.Name)

	q, err := conn.GetGroupQuery(&resourcegroups.GetGroupQueryInput{
		GroupName: aws.String(d.Id()),
	})

	isConfigurationGroup := false
	if err != nil {
		if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeBadRequestException) {
			// Attempting to get the query on a configuration group returns BadRequestException
			isConfigurationGroup = true
		} else {
			return fmt.Errorf("error reading resource query for resource group (%s): %s", d.Id(), err)
		}
	}

	if !isConfigurationGroup {
		resultQuery := map[string]interface{}{}
		resultQuery["query"] = aws.StringValue(q.GroupQuery.ResourceQuery.Query)
		resultQuery["type"] = aws.StringValue(q.GroupQuery.ResourceQuery.Type)
		if err := d.Set("resource_query", []map[string]interface{}{resultQuery}); err != nil {
			return fmt.Errorf("error setting resource_query: %s", err)
		}
	}

	if isConfigurationGroup {
		groupCfg, err := conn.GetGroupConfiguration(&resourcegroups.GetGroupConfigurationInput{
			Group: aws.String(d.Id()),
		})

		if err != nil {
			return fmt.Errorf("error reading configuration for resource group (%s): %s", d.Id(), err)
		}

		if err := d.Set("configuration", flattenResourceGroupConfigurationItems(groupCfg.GroupConfiguration.Configuration)); err != nil {
			return fmt.Errorf("error setting configuration: %s", err)
		}
	}

	tags, err := ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn

	// Conversion between a resource-query and configuration group is not possible and vice-versa
	if d.HasChange("configuration") && d.HasChange("resource_query") {
		return fmt.Errorf("conversion between resource-query and configuration group types is not possible")
	}

	if d.HasChange("description") {
		input := resourcegroups.UpdateGroupInput{
			GroupName:   aws.String(d.Id()),
			Description: aws.String(d.Get("description").(string)),
		}

		_, err := conn.UpdateGroup(&input)
		if err != nil {
			return fmt.Errorf("error updating resource group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("resource_query") {
		input := resourcegroups.UpdateGroupQueryInput{
			GroupName:     aws.String(d.Id()),
			ResourceQuery: extractResourceGroupResourceQuery(d.Get("resource_query").([]interface{})),
		}

		_, err := conn.UpdateGroupQuery(&input)
		if err != nil {
			return fmt.Errorf("error updating resource query for resource group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("configuration") {
		input := resourcegroups.PutGroupConfigurationInput{
			Group:         aws.String(d.Id()),
			Configuration: extractResourceGroupConfigurationItems(d.Get("configuration").(*schema.Set).List()),
		}

		_, err := conn.PutGroupConfiguration(&input)
		if err != nil {
			return fmt.Errorf("error updating configuration for resource group (%s): %s", d.Id(), err)
		}

		err = waitForConfigurationUpdatedState(conn, d.Id(), groupConfigurationAttachedTimeout)
		if err != nil {
			return fmt.Errorf("error updating configuration on resource group: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceGroupRead(d, meta)
}

func resourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn

	input := resourcegroups.DeleteGroupInput{
		GroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteGroup(&input)
	if err != nil {
		return fmt.Errorf("error deleting resource group (%s): %s", d.Id(), err)
	}

	return nil
}

func FindGroupByName(conn *resourcegroups.ResourceGroups, name string) (*resourcegroups.Group, error) {
	input := &resourcegroups.GetGroupInput{
		GroupName: aws.String(name),
	}

	output, err := conn.GetGroup(input)

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

func getGroupConfiguration(conn *resourcegroups.ResourceGroups, input *resourcegroups.GetGroupConfigurationInput) (*resourcegroups.GetGroupConfigurationOutput, error) {
	output, err := conn.GetGroupConfiguration(input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeNotFoundException) {
			return nil, tfresource.NewEmptyResultError(input)
		}
		return nil, err
	}

	return output, nil
}

func statusGroupConfigurationState(conn *resourcegroups.ResourceGroups, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &resourcegroups.GetGroupConfigurationInput{
			Group: aws.String(name),
		}

		output, err := getGroupConfiguration(conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.GroupConfiguration.Status), nil
	}
}

func waitForConfigurationUpdatedState(conn *resourcegroups.ResourceGroups, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			resourcegroups.GroupConfigurationStatusUpdating,
		},
		Target: []string{
			resourcegroups.GroupConfigurationStatusUpdateComplete,
			resourcegroups.GroupConfigurationStatusUpdateFailed,
		},
		Refresh: statusGroupConfigurationState(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*resourcegroups.GroupConfiguration); ok {
		return err
	}

	return err
}
