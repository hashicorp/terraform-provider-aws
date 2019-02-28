package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsResourceGroupsGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsResourceGroupsGroupCreate,
		Read:   resourceAwsResourceGroupsGroupRead,
		Update: resourceAwsResourceGroupsGroupUpdate,
		Delete: resourceAwsResourceGroupsGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"resource_query": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
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

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func diffTagsResourceGroups(oldTags map[string]interface{}, newTags map[string]interface{}) (map[string]*string, []*string) {
	create := make(map[string]*string)
	for k, v := range newTags {
		x := v.(string)
		create[k] = &x
	}

	var remove []*string
	for k, v := range oldTags {
		if _, ok := create[k]; !ok {
			r := v.(string)
			remove = append(remove, &r)
		}
	}

	return create, remove
}

func extractResourceGroupResourceQuery(resourceQueryList []interface{}) *resourcegroups.ResourceQuery {
	resourceQuery := resourceQueryList[0].(map[string]interface{})

	return &resourcegroups.ResourceQuery{
		Query: aws.String(resourceQuery["query"].(string)),
		Type:  aws.String(resourceQuery["type"].(string)),
	}
}

func resourceAwsResourceGroupsGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupsconn

	name := d.Get("name").(string)
	input := resourcegroups.CreateGroupInput{
		Description:   aws.String(d.Get("description").(string)),
		Name:          aws.String(name),
		ResourceQuery: extractResourceGroupResourceQuery(d.Get("resource_query").([]interface{})),
	}

	if t, ok := d.GetOk("tags"); ok {
		tags := make(map[string]*string)

		for k, v := range t.(map[string]interface{}) {
			tags[k] = aws.String(v.(string))
		}

		if len(tags) > 0 {
			input.Tags = tags
		}
	}

	_, err := conn.CreateGroup(&input)
	if err != nil {
		return fmt.Errorf("error creating resource group: %s", err)
	}

	d.SetId(name)

	return resourceAwsResourceGroupsGroupRead(d, meta)
}

func resourceAwsResourceGroupsGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupsconn

	g, err := conn.GetGroup(&resourcegroups.GetGroupInput{
		GroupName: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, resourcegroups.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Resource Groups Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading resource group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(g.Group.GroupArn)
	d.Set("name", aws.StringValue(g.Group.Name))
	d.Set("description", aws.StringValue(g.Group.Description))
	d.Set("arn", arn)

	q, err := conn.GetGroupQuery(&resourcegroups.GetGroupQueryInput{
		GroupName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error reading resource query for resource group (%s): %s", d.Id(), err)
	}

	resultQuery := map[string]interface{}{}
	resultQuery["query"] = aws.StringValue(q.GroupQuery.ResourceQuery.Query)
	resultQuery["type"] = aws.StringValue(q.GroupQuery.ResourceQuery.Type)
	if err := d.Set("resource_query", []map[string]interface{}{resultQuery}); err != nil {
		return fmt.Errorf("error setting resource_query: %s", err)
	}

	t, err := conn.GetTags(&resourcegroups.GetTagsInput{
		Arn: aws.String(arn),
	})

	if err != nil {
		return fmt.Errorf("error reading tags for resource group (%s): %s", d.Id(), err)
	}

	tags := make(map[string]*string)
	for k, v := range t.Tags {
		x := *v
		tags[k] = aws.String(x)
	}

	d.Set("tags", tags)

	return nil
}

func resourceAwsResourceGroupsGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupsconn

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

	if d.HasChange("tags") {
		arn := d.Get("arn")
		old, new := d.GetChange("tags")
		create, remove := diffTagsResourceGroups(old.(map[string]interface{}), new.(map[string]interface{}))

		_, err := conn.Untag(&resourcegroups.UntagInput{
			Arn:  aws.String(arn.(string)),
			Keys: remove,
		})

		if err != nil {
			return fmt.Errorf("error removing tags for resource group (%s): %s", d.Id(), err)
		}

		_, err = conn.Tag(&resourcegroups.TagInput{
			Arn:  aws.String(arn.(string)),
			Tags: create,
		})

		if err != nil {
			return fmt.Errorf("error updating tags for resource group (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsResourceGroupsGroupRead(d, meta)
}

func resourceAwsResourceGroupsGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupsconn

	input := resourcegroups.DeleteGroupInput{
		GroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteGroup(&input)
	if err != nil {
		return fmt.Errorf("error deleting resource group (%s): %s", d.Id(), err)
	}

	return nil
}
