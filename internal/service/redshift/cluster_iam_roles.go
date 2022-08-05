package redshift

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClusterIAMRoles() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterIAMRolesCreate,
		Read:   resourceClusterIAMRolesRead,
		Update: resourceClusterIAMRolesUpdate,
		Delete: resourceClusterIAMRolesDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
			Update: schema.DefaultTimeout(75 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"iam_role_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
		},
	}
}

func resourceClusterIAMRolesCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	clusterID := d.Get("cluster_identifier").(string)
	input := &redshift.ModifyClusterIamRolesInput{
		ClusterIdentifier: aws.String(clusterID),
	}

	if v, ok := d.GetOk("default_iam_role_arn"); ok {
		input.DefaultIamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_role_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.AddIamRoles = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Adding Redshift Cluster IAM Roles: %s", input)
	output, err := conn.ModifyClusterIamRoles(input)

	if err != nil {
		return fmt.Errorf("creating Redshift Cluster IAM Roles (%s): %w", clusterID, err)
	}

	d.SetId(aws.StringValue(output.Cluster.ClusterIdentifier))

	if _, err := waitClusterUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("waiting for Redshift Cluster IAM Roles (%s) update: %w", d.Id(), err)
	}

	return resourceClusterIAMRolesRead(d, meta)
}

func resourceClusterIAMRolesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	rsc, err := FindClusterByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Cluster IAM Roles (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Redshift Cluster IAM Roles (%s): %w", d.Id(), err)
	}

	var roleARNs []*string

	for _, iamRole := range rsc.IamRoles {
		roleARNs = append(roleARNs, iamRole.IamRoleArn)
	}

	d.Set("cluster_identifier", rsc.ClusterIdentifier)
	d.Set("default_iam_role_arn", rsc.DefaultIamRoleArn)
	d.Set("iam_role_arns", aws.StringValueSlice(roleARNs))

	return nil
}

func resourceClusterIAMRolesUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	o, n := d.GetChange("iam_role_arns")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	add := ns.Difference(os)
	del := os.Difference(ns)

	input := &redshift.ModifyClusterIamRolesInput{
		AddIamRoles:       flex.ExpandStringSet(add),
		ClusterIdentifier: aws.String(d.Id()),
		RemoveIamRoles:    flex.ExpandStringSet(del),
		DefaultIamRoleArn: aws.String(d.Get("default_iam_role_arn").(string)),
	}

	log.Printf("[DEBUG] Modifying Redshift Cluster IAM Roles: %s", input)
	_, err := conn.ModifyClusterIamRoles(input)

	if err != nil {
		return fmt.Errorf("updating Redshift Cluster IAM Roles (%s): %w", d.Id(), err)
	}

	if _, err := waitClusterUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("waiting for Redshift Cluster IAM Roles (%s) update: %w", d.Id(), err)
	}

	return resourceClusterIAMRolesRead(d, meta)
}

func resourceClusterIAMRolesDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	input := &redshift.ModifyClusterIamRolesInput{
		ClusterIdentifier: aws.String(d.Id()),
		RemoveIamRoles:    flex.ExpandStringSet(d.Get("iam_role_arns").(*schema.Set)),
		DefaultIamRoleArn: aws.String(d.Get("default_iam_role_arn").(string)),
	}

	log.Printf("[DEBUG] Removing Redshift Cluster IAM Roles: %s", input)
	_, err := conn.ModifyClusterIamRoles(input)

	if err != nil {
		return fmt.Errorf("deleting Redshift Cluster IAM Roles (%s): %w", d.Id(), err)
	}

	if _, err := waitClusterUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("waiting for Redshift Cluster IAM Roles (%s) update: %w", d.Id(), err)
	}

	return nil
}
