package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsWorkspaceBundle() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWorkspaceBundleRead,

		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_type": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"user_storage": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"root_storage": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsWorkspaceBundleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wsconn

	bundleID := d.Get("bundle_id").(string)
	input := &workspaces.DescribeWorkspaceBundlesInput{
		BundleIds: []*string{aws.String(bundleID)},
	}

	resp, err := conn.DescribeWorkspaceBundles(input)
	if err != nil {
		return err
	}

	if len(resp.Bundles) != 1 {
		return fmt.Errorf("The number of Workspace Bundle (%s) should be 1, but %d", bundleID, len(resp.Bundles))
	}

	bundle := resp.Bundles[0]
	d.SetId(bundleID)
	d.Set("description", bundle.Description)
	d.Set("name", bundle.Name)
	d.Set("owner", bundle.Owner)
	if bundle.ComputeType != nil {
		r := map[string]interface{}{
			"name": *bundle.ComputeType.Name,
		}
		ct := []map[string]interface{}{r}
		d.Set("compute_type", ct)
	}
	if bundle.RootStorage != nil {
		r := map[string]interface{}{
			"capacity": *bundle.RootStorage.Capacity,
		}
		rs := []map[string]interface{}{r}
		d.Set("root_storage", rs)
	}
	if bundle.UserStorage != nil {
		r := map[string]interface{}{
			"capacity": *bundle.UserStorage.Capacity,
		}
		us := []map[string]interface{}{r}
		d.Set("user_storage", us)
	}

	return nil
}
