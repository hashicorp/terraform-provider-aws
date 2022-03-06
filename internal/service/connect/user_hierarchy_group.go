package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceUserHierarchyGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceUserHierarchyGroupRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hierarchy_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hierarchy_path": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"level_one": func() *schema.Schema {
							schema := connectUserHierarchyPathLevelSchema()
							return schema
						}(),
						"level_two": func() *schema.Schema {
							schema := connectUserHierarchyPathLevelSchema()
							return schema
						}(),
						"level_three": func() *schema.Schema {
							schema := connectUserHierarchyPathLevelSchema()
							return schema
						}(),
						"level_four": func() *schema.Schema {
							schema := connectUserHierarchyPathLevelSchema()
							return schema
						}(),
						"level_five": func() *schema.Schema {
							schema := connectUserHierarchyPathLevelSchema()
							return schema
						}(),
					},
				},
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"level_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"parent_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

// Each level shares the same schema
func connectUserHierarchyPathLevelSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func resourceUserHierarchyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID, userHierarchyGroupID, err := UserHierarchyGroupParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeUserHierarchyGroupWithContext(ctx, &connect.DescribeUserHierarchyGroupInput{
		HierarchyGroupId: aws.String(userHierarchyGroupID),
		InstanceId:       aws.String(instanceID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect User Hierarchy Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect User Hierarchy Group (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.HierarchyGroup == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect User Hierarchy Group (%s): empty response", d.Id()))
	}

	d.Set("arn", resp.HierarchyGroup.Arn)
	d.Set("hierarchy_group_id", resp.HierarchyGroup.Id)
	d.Set("instance_id", instanceID)
	d.Set("level_id", resp.HierarchyGroup.LevelId)
	d.Set("name", resp.HierarchyGroup.Name)

	if err := d.Set("hierarchy_path", flattenUserHierarchyPath(resp.HierarchyGroup.HierarchyPath)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting Connect User Hierarchy Group hierarchy_path (%s): %w", d.Id(), err))
	}

	tags := KeyValueTags(resp.HierarchyGroup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func UserHierarchyGroupParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:userHierarchyGroupID", id)
	}

	return parts[0], parts[1], nil
}

func flattenUserHierarchyPath(userHierarchyPath *connect.HierarchyPath) []interface{} {
	if userHierarchyPath == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if userHierarchyPath.LevelOne != nil {
		values["level_one"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelOne)
	}

	if userHierarchyPath.LevelTwo != nil {
		values["level_two"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelTwo)
	}

	if userHierarchyPath.LevelThree != nil {
		values["level_three"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelThree)
	}

	if userHierarchyPath.LevelFour != nil {
		values["level_four"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelFour)
	}

	if userHierarchyPath.LevelFive != nil {
		values["level_five"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelFive)
	}

	return []interface{}{values}
}

func flattenUserHierarchyPathLevel(userHierarchyPathLevel *connect.HierarchyGroupSummary) []interface{} {
	if userHierarchyPathLevel == nil {
		return []interface{}{}
	}

	level := map[string]interface{}{
		"arn":  aws.StringValue(userHierarchyPathLevel.Arn),
		"id":   aws.StringValue(userHierarchyPathLevel.Id),
		"name": aws.StringValue(userHierarchyPathLevel.Name),
	}

	return []interface{}{level}
}
