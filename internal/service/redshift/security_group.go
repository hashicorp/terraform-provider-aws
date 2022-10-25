package redshift

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

func ResourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceSecurityGroupCreate,
		Read:   resourceSecurityGroupRead,
		Update: resourceSecurityGroupUpdate,
		Delete: resourceSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringNotInSlice([]string{"default"}, false),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
				),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},

			"ingress": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"security_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"security_group_owner_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Set: resourceSecurityGroupIngressHash,
			},
		},

		DeprecationMessage: `With the retirement of EC2-Classic the aws_redshift_security_group resource has been deprecated and will be removed in a future version.`,
	}
}

func resourceSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	return errors.New(`with the retirement of EC2-Classic no new Redshift Security Groups can be created`)
}

func resourceSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	sg, err := resourceSecurityGroupRetrieve(d, meta)
	if err != nil {
		return err
	}

	rules := &schema.Set{
		F: resourceSecurityGroupIngressHash,
	}

	for _, v := range sg.IPRanges {
		rule := map[string]interface{}{"cidr": aws.StringValue(v.CIDRIP)}
		rules.Add(rule)
	}

	for _, g := range sg.EC2SecurityGroups {
		rule := map[string]interface{}{
			"security_group_name":     aws.StringValue(g.EC2SecurityGroupName),
			"security_group_owner_id": aws.StringValue(g.EC2SecurityGroupOwnerId),
		}
		rules.Add(rule)
	}

	d.Set("ingress", rules)
	d.Set("name", sg.ClusterSecurityGroupName)
	d.Set("description", sg.Description)

	return nil
}

func resourceSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.HasChange("ingress") {
		o, n := d.GetChange("ingress")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removeIngressRules := expandSGRevokeIngress(os.Difference(ns).List())
		if len(removeIngressRules) > 0 {
			for _, r := range removeIngressRules {
				r.ClusterSecurityGroupName = aws.String(d.Id())

				_, err := conn.RevokeClusterSecurityGroupIngress(&r)
				if err != nil {
					return err
				}
			}
		}

		addIngressRules := expandSGAuthorizeIngress(ns.Difference(os).List())
		if len(addIngressRules) > 0 {
			for _, r := range addIngressRules {
				r.ClusterSecurityGroupName = aws.String(d.Id())

				_, err := conn.AuthorizeClusterSecurityGroupIngress(&r)
				if err != nil {
					return err
				}
			}
		}
	}
	return resourceSecurityGroupRead(d, meta)
}

func resourceSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	log.Printf("[DEBUG] Redshift Security Group destroy: %v", d.Id())
	opts := redshift.DeleteClusterSecurityGroupInput{
		ClusterSecurityGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Redshift Security Group destroy configuration: %v", opts)
	_, err := conn.DeleteClusterSecurityGroup(&opts)

	if err != nil {
		newerr, ok := err.(awserr.Error)
		if ok && newerr.Code() == "InvalidRedshiftSecurityGroup.NotFound" {
			return nil
		}
		return err
	}

	return nil
}

func resourceSecurityGroupRetrieve(d *schema.ResourceData, meta interface{}) (*redshift.ClusterSecurityGroup, error) {
	conn := meta.(*conns.AWSClient).RedshiftConn

	opts := redshift.DescribeClusterSecurityGroupsInput{
		ClusterSecurityGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Redshift Security Group describe configuration: %#v", opts)

	resp, err := conn.DescribeClusterSecurityGroups(&opts)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving Redshift Security Groups: %s", err)
	}

	if len(resp.ClusterSecurityGroups) != 1 ||
		aws.StringValue(resp.ClusterSecurityGroups[0].ClusterSecurityGroupName) != d.Id() {
		return nil, fmt.Errorf("Unable to find Redshift Security Group: %#v", resp.ClusterSecurityGroups)
	}

	return resp.ClusterSecurityGroups[0], nil
}

func resourceSecurityGroupIngressHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["cidr"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["security_group_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["security_group_owner_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

func expandSGAuthorizeIngress(configured []interface{}) []redshift.AuthorizeClusterSecurityGroupIngressInput {
	var ingress []redshift.AuthorizeClusterSecurityGroupIngressInput

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		i := redshift.AuthorizeClusterSecurityGroupIngressInput{}

		if v, ok := data["cidr"]; ok {
			i.CIDRIP = aws.String(v.(string))
		}

		if v, ok := data["security_group_name"]; ok {
			i.EC2SecurityGroupName = aws.String(v.(string))
		}

		if v, ok := data["security_group_owner_id"]; ok {
			i.EC2SecurityGroupOwnerId = aws.String(v.(string))
		}

		ingress = append(ingress, i)
	}

	return ingress
}

func expandSGRevokeIngress(configured []interface{}) []redshift.RevokeClusterSecurityGroupIngressInput {
	var ingress []redshift.RevokeClusterSecurityGroupIngressInput

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		i := redshift.RevokeClusterSecurityGroupIngressInput{}

		if v, ok := data["cidr"]; ok {
			i.CIDRIP = aws.String(v.(string))
		}

		if v, ok := data["security_group_name"]; ok {
			i.EC2SecurityGroupName = aws.String(v.(string))
		}

		if v, ok := data["security_group_owner_id"]; ok {
			i.EC2SecurityGroupOwnerId = aws.String(v.(string))
		}

		ingress = append(ingress, i)
	}

	return ingress
}
