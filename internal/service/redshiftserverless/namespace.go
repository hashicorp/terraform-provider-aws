// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshiftserverless_namespace", name="Namespace")
// @Tags(identifierAttribute="arn")
func resourceNamespace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNamespaceCreate,
		ReadWithoutTimeout:   resourceNamespaceRead,
		UpdateWithoutTimeout: resourceNamespaceUpdate,
		DeleteWithoutTimeout: resourceNamespaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"admin_password_secret_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_password_secret_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidKMSKeyID,
			},
			"admin_user_password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"manage_admin_password"},
			},
			"admin_username": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Computed:  true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"default_iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"iam_roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"log_exports": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(redshiftserverless.LogExport_Values(), false),
				},
			},
			"manage_admin_password": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"admin_user_password"},
			},
			"namespace_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	name := d.Get("namespace_name").(string)
	input := &redshiftserverless.CreateNamespaceInput{
		NamespaceName: aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("admin_password_secret_kms_key_id"); ok {
		input.AdminPasswordSecretKmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("admin_user_password"); ok {
		input.AdminUserPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("admin_username"); ok {
		input.AdminUsername = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_name"); ok {
		input.DbName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_iam_role_arn"); ok {
		input.DefaultIamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_roles"); ok && v.(*schema.Set).Len() > 0 {
		input.IamRoles = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_exports"); ok && v.(*schema.Set).Len() > 0 {
		input.LogExports = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("manage_admin_password"); ok {
		input.ManageAdminPassword = aws.Bool(v.(bool))
	}

	output, err := conn.CreateNamespaceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Namespace (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Namespace.NamespaceName))

	return append(diags, resourceNamespaceRead(ctx, d, meta)...)
}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	output, err := findNamespaceByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Namespace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Namespace (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(output.NamespaceArn)
	d.Set("admin_password_secret_arn", output.AdminPasswordSecretArn)
	d.Set("admin_password_secret_kms_key_id", output.AdminPasswordSecretKmsKeyId)
	d.Set("admin_username", output.AdminUsername)
	d.Set(names.AttrARN, arn)
	d.Set("db_name", output.DbName)
	d.Set("default_iam_role_arn", output.DefaultIamRoleArn)
	d.Set("iam_roles", flattenNamespaceIAMRoles(output.IamRoles))
	d.Set(names.AttrKMSKeyID, output.KmsKeyId)
	d.Set("log_exports", aws.StringValueSlice(output.LogExports))
	d.Set("namespace_id", output.NamespaceId)
	d.Set("namespace_name", output.NamespaceName)

	return diags
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &redshiftserverless.UpdateNamespaceInput{
			NamespaceName: aws.String(d.Id()),
		}

		if d.HasChanges("admin_password_secret_kms_key_id") {
			input.AdminPasswordSecretKmsKeyId = aws.String(d.Get("admin_password_secret_kms_key_id").(string))
		}

		if d.HasChanges("admin_username", "admin_user_password") {
			input.AdminUsername = aws.String(d.Get("admin_username").(string))
			input.AdminUserPassword = aws.String(d.Get("admin_user_password").(string))
		}

		if d.HasChange("default_iam_role_arn") {
			input.DefaultIamRoleArn = aws.String(d.Get("default_iam_role_arn").(string))
		}

		if d.HasChange("iam_roles") {
			input.IamRoles = flex.ExpandStringSet(d.Get("iam_roles").(*schema.Set))
		}

		if d.HasChange(names.AttrKMSKeyID) {
			input.KmsKeyId = aws.String(d.Get(names.AttrKMSKeyID).(string))
		}

		if d.HasChange("log_exports") {
			input.LogExports = flex.ExpandStringSet(d.Get("log_exports").(*schema.Set))
		}

		if d.HasChange("manage_admin_password") {
			input.ManageAdminPassword = aws.Bool(d.Get("manage_admin_password").(bool))
		}

		_, err := conn.UpdateNamespaceWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Namespace (%s): %s", d.Id(), err)
		}

		if _, err := waitNamespaceUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Namespace (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceNamespaceRead(ctx, d, meta)...)
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Namespace: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 10*time.Minute,
		func() (interface{}, error) {
			return conn.DeleteNamespaceWithContext(ctx, &redshiftserverless.DeleteNamespaceInput{
				NamespaceName: aws.String(d.Id()),
			})
		},
		// "ConflictException: There is an operation running on the namespace. Try deleting the namespace again later."
		redshiftserverless.ErrCodeConflictException, "operation running")

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Namespace (%s): %s", d.Id(), err)
	}

	if _, err := waitNamespaceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Namespace (%s) delete: %s", d.Id(), err)
	}

	return diags
}

var (
	reIAMRole = regexache.MustCompile(`^\s*IamRole\((.*)\)\s*$`)
)

func flattenNamespaceIAMRoles(iamRoles []*string) []string {
	var tfList []string

	for _, iamRole := range iamRoles {
		iamRole := aws.StringValue(iamRole)

		if arn.IsARN(iamRole) {
			tfList = append(tfList, iamRole)
			continue
		}

		// e.g. "IamRole(applyStatus=in-sync, iamRoleArn=arn:aws:iam::123456789012:role/service-role/test)"
		if m := reIAMRole.FindStringSubmatch(iamRole); len(m) > 0 {
			var key string
			s := m[1]
			for s != "" {
				key, s, _ = strings.Cut(s, ",")
				key = strings.TrimSpace(key)
				if key == "" {
					continue
				}
				key, value, _ := strings.Cut(key, "=")
				if key == "iamRoleArn" {
					tfList = append(tfList, value)
					break
				}
			}

			continue
		}
	}

	return tfList
}
