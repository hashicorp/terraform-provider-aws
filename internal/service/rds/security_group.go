package rds

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

						"security_group_id": {
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

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,

		DeprecationMessage: `With the retirement of EC2-Classic the aws_db_security_group resource has been deprecated and will be removed in a future version.`,
	}
}

func resourceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return sdkdiag.AppendErrorf(diags, `with the retirement of EC2-Classic no new RDS DB Security Groups can be created`)
}

func resourceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	sg, err := resourceSecurityGroupRetrieve(ctx, d, meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Security Group (%s): %s", d.Id(), err)
	}

	d.Set("name", sg.DBSecurityGroupName)
	d.Set("description", sg.DBSecurityGroupDescription)

	// Create an empty schema.Set to hold all ingress rules
	rules := &schema.Set{
		F: resourceSecurityGroupIngressHash,
	}

	for _, v := range sg.IPRanges {
		rule := map[string]interface{}{"cidr": *v.CIDRIP}
		rules.Add(rule)
	}

	for _, g := range sg.EC2SecurityGroups {
		rule := map[string]interface{}{}
		if g.EC2SecurityGroupId != nil {
			rule["security_group_id"] = aws.StringValue(g.EC2SecurityGroupId)
		}
		if g.EC2SecurityGroupName != nil {
			rule["security_group_name"] = aws.StringValue(g.EC2SecurityGroupName)
		}
		if g.EC2SecurityGroupOwnerId != nil {
			rule["security_group_owner_id"] = aws.StringValue(g.EC2SecurityGroupOwnerId)
		}
		rules.Add(rule)
	}

	d.Set("ingress", rules)

	conn := meta.(*conns.AWSClient).RDSConn()

	arn := aws.StringValue(sg.DBSecurityGroupArn)
	d.Set("arn", arn)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for RDS DB Security Group (%s): %s", d.Get("arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Security Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	if d.HasChange("ingress") {
		sg, err := resourceSecurityGroupRetrieve(ctx, d, meta)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Security Group (%s): %s", d.Id(), err)
		}

		oi, ni := d.GetChange("ingress")
		if oi == nil {
			oi = new(schema.Set)
		}
		if ni == nil {
			ni = new(schema.Set)
		}

		ois := oi.(*schema.Set)
		nis := ni.(*schema.Set)
		removeIngress := ois.Difference(nis).List()
		newIngress := nis.Difference(ois).List()

		// DELETE old Ingress rules
		for _, ing := range removeIngress {
			err := resourceSecurityGroupRevokeRule(ctx, ing, *sg.DBSecurityGroupName, conn)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Security Group (%s): revoking ingress: %s", d.Id(), err)
			}
		}

		// ADD new/updated Ingress rules
		for _, ing := range newIngress {
			err := resourceSecurityGroupAuthorizeRule(ctx, ing, *sg.DBSecurityGroupName, conn)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Security Group (%s): authorizing ingress: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceSecurityGroupRead(ctx, d, meta)...)
}

func resourceSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	log.Printf("[DEBUG] DB Security Group destroy: %v", d.Id())

	opts := rds.DeleteDBSecurityGroupInput{DBSecurityGroupName: aws.String(d.Id())}

	_, err := conn.DeleteDBSecurityGroupWithContext(ctx, &opts)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, "InvalidDBSecurityGroup.NotFound") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Security Group (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSecurityGroupRetrieve(ctx context.Context, d *schema.ResourceData, meta interface{}) (*rds.DBSecurityGroup, error) {
	conn := meta.(*conns.AWSClient).RDSConn()

	opts := rds.DescribeDBSecurityGroupsInput{
		DBSecurityGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] DB Security Group describe configuration: %#v", opts)

	resp, err := conn.DescribeDBSecurityGroupsWithContext(ctx, &opts)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving DB Security Groups: %s", err)
	}

	if len(resp.DBSecurityGroups) != 1 ||
		aws.StringValue(resp.DBSecurityGroups[0].DBSecurityGroupName) != d.Id() {
		return nil, fmt.Errorf("Unable to find DB Security Group: %#v", resp.DBSecurityGroups)
	}

	return resp.DBSecurityGroups[0], nil
}

// Authorizes the ingress rule on the db security group
func resourceSecurityGroupAuthorizeRule(ctx context.Context, ingress interface{}, dbSecurityGroupName string, conn *rds.RDS) error {
	ing := ingress.(map[string]interface{})

	opts := rds.AuthorizeDBSecurityGroupIngressInput{
		DBSecurityGroupName: aws.String(dbSecurityGroupName),
	}

	if attr, ok := ing["cidr"]; ok && attr != "" {
		opts.CIDRIP = aws.String(attr.(string))
	}

	if attr, ok := ing["security_group_name"]; ok && attr != "" {
		opts.EC2SecurityGroupName = aws.String(attr.(string))
	}

	if attr, ok := ing["security_group_id"]; ok && attr != "" {
		opts.EC2SecurityGroupId = aws.String(attr.(string))
	}

	if attr, ok := ing["security_group_owner_id"]; ok && attr != "" {
		opts.EC2SecurityGroupOwnerId = aws.String(attr.(string))
	}

	log.Printf("[DEBUG] Authorize ingress rule configuration: %#v", opts)

	_, err := conn.AuthorizeDBSecurityGroupIngressWithContext(ctx, &opts)

	if err != nil {
		return fmt.Errorf("Error authorizing security group ingress: %s", err)
	}

	return nil
}

// Revokes the ingress rule on the db security group
func resourceSecurityGroupRevokeRule(ctx context.Context, ingress interface{}, dbSecurityGroupName string, conn *rds.RDS) error {
	ing := ingress.(map[string]interface{})

	opts := rds.RevokeDBSecurityGroupIngressInput{
		DBSecurityGroupName: aws.String(dbSecurityGroupName),
	}

	if attr, ok := ing["cidr"]; ok && attr != "" {
		opts.CIDRIP = aws.String(attr.(string))
	}

	if attr, ok := ing["security_group_name"]; ok && attr != "" {
		opts.EC2SecurityGroupName = aws.String(attr.(string))
	}

	if attr, ok := ing["security_group_id"]; ok && attr != "" {
		opts.EC2SecurityGroupId = aws.String(attr.(string))
	}

	if attr, ok := ing["security_group_owner_id"]; ok && attr != "" {
		opts.EC2SecurityGroupOwnerId = aws.String(attr.(string))
	}

	log.Printf("[DEBUG] Revoking ingress rule configuration: %#v", opts)

	_, err := conn.RevokeDBSecurityGroupIngressWithContext(ctx, &opts)

	if err != nil {
		return fmt.Errorf("Error revoking security group ingress: %s", err)
	}

	return nil
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

	if v, ok := m["security_group_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["security_group_owner_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}
