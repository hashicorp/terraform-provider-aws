package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceUserHierarchyStructure() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceUserHierarchyStructureRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"hierarchy_structure": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"level_one": {
							Type:     schema.TypeList,
							Optional: true,
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
										Required: true,
									},
								},
							},
						},
						"level_two": {
							Type:     schema.TypeList,
							Optional: true,
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
										Required: true,
									},
								},
							},
						},
						"level_three": {
							Type:     schema.TypeList,
							Optional: true,
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
										Required: true,
									},
								},
							},
						},
						"level_four": {
							Type:     schema.TypeList,
							Optional: true,
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
										Required: true,
									},
								},
							},
						},
						"level_five": {
							Type:     schema.TypeList,
							Optional: true,
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
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
		},
	}
}

func resourceUserHierarchyStructureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, _, err := UserHierarchyStructureParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeUserHierarchyStructureWithContext(ctx, &connect.DescribeUserHierarchyStructureInput{
		InstanceId: aws.String(instanceID),
	})

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, connect.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Connect User Hierarchy Structure (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect User Hierarchy Structure (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.HierarchyStructure == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect User Hierarchy Structure (%s): empty response", d.Id()))
	}

	if err := d.Set("hierarchy_structure", flattenUserHierarchyStructure(resp.HierarchyStructure)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting Connect User Hierarchy Structure hierarchy_structure for Connect instance: (%s)", d.Id()))
	}

	d.Set("instance_id", instanceID)

	return nil
}

func flattenUserHierarchyStructure(userHierarchyStructure *connect.HierarchyStructure) []interface{} {
	if userHierarchyStructure == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if userHierarchyStructure.LevelOne != nil {
		values["level_one"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelOne)
	}

	if userHierarchyStructure.LevelTwo != nil {
		values["level_two"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelTwo)
	}

	if userHierarchyStructure.LevelThree != nil {
		values["level_three"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelThree)
	}

	if userHierarchyStructure.LevelFour != nil {
		values["level_four"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelFour)
	}

	if userHierarchyStructure.LevelFive != nil {
		values["level_five"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelFive)
	}

	return []interface{}{values}
}

func flattenUserHierarchyStructureLevel(userHierarchyStructureLevel *connect.HierarchyLevel) []interface{} {
	if userHierarchyStructureLevel == nil {
		return []interface{}{}
	}

	level := map[string]interface{}{
		"arn":  aws.StringValue(userHierarchyStructureLevel.Arn),
		"id":   aws.StringValue(userHierarchyStructureLevel.Id),
		"name": aws.StringValue(userHierarchyStructureLevel.Name),
	}

	return []interface{}{level}
}

func UserHierarchyStructureParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:hierarchyStructureID", id)
	}

	return parts[0], parts[1], nil
}
