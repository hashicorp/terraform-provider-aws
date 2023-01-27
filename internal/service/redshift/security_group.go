package redshift

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityGroupCreate,
		ReadWithoutTimeout:   resourceSecurityGroupRead,
		UpdateWithoutTimeout: resourceSecurityGroupUpdate,
		DeleteWithoutTimeout: resourceSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return sdkdiag.AppendErrorf(diags, `with the retirement of EC2-Classic no new Redshift Security Groups can be created`)
}

func resourceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	sg, err := resourceSecurityGroupRetrieve(ctx, d, meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Security Group (%s): %s", d.Id(), err)
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

	return diags
}

func resourceSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()

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

				_, err := conn.RevokeClusterSecurityGroupIngressWithContext(ctx, &r)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Redshift Security Group (%s): revoking ingress: %s", d.Id(), err)
				}
			}
		}

		addIngressRules := expandSGAuthorizeIngress(ns.Difference(os).List())
		if len(addIngressRules) > 0 {
			for _, r := range addIngressRules {
				r.ClusterSecurityGroupName = aws.String(d.Id())

				_, err := conn.AuthorizeClusterSecurityGroupIngressWithContext(ctx, &r)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Redshift Security Group (%s): authorizing ingress: %s", d.Id(), err)
				}
			}
		}
	}
	return append(diags, resourceSecurityGroupRead(ctx, d, meta)...)
}

func resourceSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()

	log.Printf("[DEBUG] Redshift Security Group destroy: %v", d.Id())
	opts := redshift.DeleteClusterSecurityGroupInput{
		ClusterSecurityGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteClusterSecurityGroupWithContext(ctx, &opts)

	if err != nil {
		newerr, ok := err.(awserr.Error)
		if ok && newerr.Code() == "InvalidRedshiftSecurityGroup.NotFound" {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Security Group (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSecurityGroupRetrieve(ctx context.Context, d *schema.ResourceData, meta interface{}) (*redshift.ClusterSecurityGroup, error) {
	conn := meta.(*conns.AWSClient).RedshiftConn()

	opts := redshift.DescribeClusterSecurityGroupsInput{
		ClusterSecurityGroupName: aws.String(d.Id()),
	}

	resp, err := conn.DescribeClusterSecurityGroupsWithContext(ctx, &opts)
	if err != nil {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: opts,
		}
	}

	if len(resp.ClusterSecurityGroups) == 0 || resp.ClusterSecurityGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(opts)
	}

	if l := len(resp.ClusterSecurityGroups); l > 1 {
		return nil, tfresource.NewTooManyResultsError(l, opts)
	}

	result := resp.ClusterSecurityGroups[0]
	if aws.StringValue(result.ClusterSecurityGroupName) != d.Id() {
		return nil, &resource.NotFoundError{
			LastRequest: opts,
		}
	}

	return result, nil
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
