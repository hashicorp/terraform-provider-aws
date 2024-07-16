// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_cluster_iam_roles", name="Cluster IAM Roles")
func resourceClusterIAMRoles() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterIAMRolesCreate,
		ReadWithoutTimeout:   resourceClusterIAMRolesRead,
		UpdateWithoutTimeout: resourceClusterIAMRolesUpdate,
		DeleteWithoutTimeout: resourceClusterIAMRolesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
			Update: schema.DefaultTimeout(75 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrClusterIdentifier: {
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

func resourceClusterIAMRolesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	clusterID := d.Get(names.AttrClusterIdentifier).(string)
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
	output, err := conn.ModifyClusterIamRolesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster IAM Roles (%s): %s", clusterID, err)
	}

	d.SetId(aws.StringValue(output.Cluster.ClusterIdentifier))

	if _, err := waitClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster IAM Roles (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceClusterIAMRolesRead(ctx, d, meta)...)
}

func resourceClusterIAMRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	rsc, err := findClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Cluster IAM Roles (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Cluster IAM Roles (%s): %s", d.Id(), err)
	}

	var roleARNs []*string

	for _, iamRole := range rsc.IamRoles {
		roleARNs = append(roleARNs, iamRole.IamRoleArn)
	}

	d.Set(names.AttrClusterIdentifier, rsc.ClusterIdentifier)
	d.Set("default_iam_role_arn", rsc.DefaultIamRoleArn)
	d.Set("iam_role_arns", aws.StringValueSlice(roleARNs))

	return diags
}

func resourceClusterIAMRolesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

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
	_, err := conn.ModifyClusterIamRolesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster IAM Roles (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster IAM Roles (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceClusterIAMRolesRead(ctx, d, meta)...)
}

func resourceClusterIAMRolesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	input := &redshift.ModifyClusterIamRolesInput{
		ClusterIdentifier: aws.String(d.Id()),
		RemoveIamRoles:    flex.ExpandStringSet(d.Get("iam_role_arns").(*schema.Set)),
		DefaultIamRoleArn: aws.String(d.Get("default_iam_role_arn").(string)),
	}

	log.Printf("[DEBUG] Removing Redshift Cluster IAM Roles: %s", input)
	_, err := conn.ModifyClusterIamRolesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Cluster IAM Roles (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster IAM Roles (%s) update: %s", d.Id(), err)
	}

	return diags
}
