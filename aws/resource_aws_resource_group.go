package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsResourceGroupCreate,
		Read:   resourceAwsResourceGroupRead,
		Update: resourceAwsResourceGroupUpdate,
		Delete: resourceAwsResourceGroupDelete,

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
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},

						"type": {
							Type:     schema.TypeString,
							Required: true,
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

func resourceAwsResourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupconn
	resourceQueryList := d.Get("resource_query").([]interface{})
	resourceQuery := resourceQueryList[0].(map[string]interface{})

	input := resourcegroups.CreateGroupInput{
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(d.Get("name").(string)),
		ResourceQuery: &resourcegroups.ResourceQuery{
			Query: aws.String(resourceQuery["query"].(string)),
			Type:  aws.String(resourceQuery["type"].(string)),
		},
	}

	group, err := conn.CreateGroup(input)
	if err != nil {
		return err
	}

	return resourceAwsResourceGroupRead(d, meta)
}

func resourceAwsResourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupconn
	return nil
}

func resourceAwsResourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupconn
	return resourceAwsResourceGroupRead(d, meta)
}

func resourceAwsResourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupconn

	return nil
}
