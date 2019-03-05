package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsAppmeshVirtualRouter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppmeshVirtualRouterCreate,
		Read:   resourceAwsAppmeshVirtualRouterRead,
		Delete: resourceAwsAppmeshVirtualRouterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshVirtualRouterImport,
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsAppmeshVirtualRouterMigrateState,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"spec": {
				Type:     schema.TypeList,
				Removed:  "Remove `spec` configuration block completely",
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppmeshVirtualRouterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	req := &appmesh.CreateVirtualRouterInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		VirtualRouterName: aws.String(d.Get("name").(string)),
		Spec:              &appmesh.VirtualRouterSpec{},
	}

	log.Printf("[DEBUG] Creating App Mesh virtual router: %#v", req)
	resp, err := conn.CreateVirtualRouter(req)
	if err != nil {
		return fmt.Errorf("error creating App Mesh virtual router: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualRouter.Metadata.Uid))

	return resourceAwsAppmeshVirtualRouterRead(d, meta)
}

func resourceAwsAppmeshVirtualRouterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	resp, err := conn.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		VirtualRouterName: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh virtual router (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading App Mesh virtual router: %s", err)
	}
	if aws.StringValue(resp.VirtualRouter.Status.Status) == appmesh.VirtualRouterStatusCodeDeleted {
		log.Printf("[WARN] App Mesh virtual router (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", resp.VirtualRouter.VirtualRouterName)
	d.Set("mesh_name", resp.VirtualRouter.MeshName)
	d.Set("arn", resp.VirtualRouter.Metadata.Arn)
	d.Set("created_date", resp.VirtualRouter.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualRouter.Metadata.LastUpdatedAt.Format(time.RFC3339))

	return nil
}

func resourceAwsAppmeshVirtualRouterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh virtual router: %s", d.Id())
	_, err := conn.DeleteVirtualRouter(&appmesh.DeleteVirtualRouterInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		VirtualRouterName: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting App Mesh virtual router: %s", err)
	}

	return nil
}

func resourceAwsAppmeshVirtualRouterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-router-name'", d.Id())
	}

	mesh := parts[0]
	name := parts[1]
	log.Printf("[DEBUG] Importing App Mesh virtual router %s from mesh %s", name, mesh)

	conn := meta.(*AWSClient).appmeshconn

	resp, err := conn.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
		MeshName:          aws.String(mesh),
		VirtualRouterName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.VirtualRouter.Metadata.Uid))
	d.Set("name", resp.VirtualRouter.VirtualRouterName)
	d.Set("mesh_name", resp.VirtualRouter.MeshName)

	return []*schema.ResourceData{d}, nil
}
