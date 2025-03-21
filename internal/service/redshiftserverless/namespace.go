// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
				ConflictsWith: []string{"manage_admin_password", "admin_user_password_wo"},
			},
			"admin_user_password_wo": {
				Type:          schema.TypeString,
				Optional:      true,
				WriteOnly:     true,
				ConflictsWith: []string{"admin_user_password", "manage_admin_password"},
				RequiredWith:  []string{"admin_user_password_wo_version"},
			},
			"admin_user_password_wo_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"admin_user_password_wo"},
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.LogExport](),
				},
			},
			"manage_admin_password": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"admin_user_password", "admin_user_password_wo"},
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
	}
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	name := d.Get("namespace_name").(string)
	input := &redshiftserverless.CreateNamespaceInput{
		NamespaceName: aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	adminUserPasswordWO, di := flex.GetWriteOnlyStringValue(d, cty.GetAttrPath("admin_user_password_wo"))
	diags = append(diags, di...)
	if diags.HasError() {
		return diags
	}

	if v, ok := d.GetOk("admin_password_secret_kms_key_id"); ok {
		input.AdminPasswordSecretKmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("admin_user_password"); ok {
		input.AdminUserPassword = aws.String(v.(string))
	}

	if adminUserPasswordWO != "" {
		input.AdminUserPassword = aws.String(adminUserPasswordWO)
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
		input.IamRoles = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_exports"); ok && v.(*schema.Set).Len() > 0 {
		input.LogExports = flex.ExpandStringyValueSet[awstypes.LogExport](v.(*schema.Set))
	}

	if v, ok := d.GetOk("manage_admin_password"); ok {
		input.ManageAdminPassword = aws.Bool(v.(bool))
	}

	output, err := conn.CreateNamespace(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Namespace (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Namespace.NamespaceName))

	return append(diags, resourceNamespaceRead(ctx, d, meta)...)
}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	output, err := findNamespaceByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Namespace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Namespace (%s): %s", d.Id(), err)
	}

	d.Set("admin_password_secret_arn", output.AdminPasswordSecretArn)
	d.Set("admin_password_secret_kms_key_id", output.AdminPasswordSecretKmsKeyId)
	d.Set("admin_username", output.AdminUsername)
	d.Set(names.AttrARN, output.NamespaceArn)
	d.Set("db_name", output.DbName)
	d.Set("default_iam_role_arn", output.DefaultIamRoleArn)
	d.Set("iam_roles", flattenNamespaceIAMRoles(output.IamRoles))
	d.Set(names.AttrKMSKeyID, output.KmsKeyId)
	d.Set("log_exports", flex.FlattenStringyValueSet[awstypes.LogExport](output.LogExports))
	d.Set("namespace_id", output.NamespaceId)
	d.Set("namespace_name", output.NamespaceName)

	return diags
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &redshiftserverless.UpdateNamespaceInput{
			NamespaceName: aws.String(d.Id()),
		}

		if d.HasChanges("admin_password_secret_kms_key_id") {
			input.AdminPasswordSecretKmsKeyId = aws.String(d.Get("admin_password_secret_kms_key_id").(string))
		}

		if d.HasChanges("admin_username", "admin_user_password", "admin_user_password_wo_version") {
			input.AdminUsername = aws.String(d.Get("admin_username").(string))

			if v, ok := d.Get("admin_user_password").(string); ok {
				input.AdminUserPassword = aws.String(v)
			}

			adminUserPasswordWO, di := flex.GetWriteOnlyStringValue(d, cty.GetAttrPath("admin_user_password_wo"))
			diags = append(diags, di...)
			if diags.HasError() {
				return diags
			}

			if adminUserPasswordWO != "" {
				input.AdminUserPassword = aws.String(adminUserPasswordWO)
			}
		}

		if d.HasChange("default_iam_role_arn") {
			input.DefaultIamRoleArn = aws.String(d.Get("default_iam_role_arn").(string))
		}

		if d.HasChange("iam_roles") {
			input.IamRoles = flex.ExpandStringValueSet(d.Get("iam_roles").(*schema.Set))
		}

		if d.HasChange(names.AttrKMSKeyID) {
			input.KmsKeyId = aws.String(d.Get(names.AttrKMSKeyID).(string))
		}

		if d.HasChange("log_exports") {
			input.LogExports = flex.ExpandStringyValueSet[awstypes.LogExport](d.Get("log_exports").(*schema.Set))
		}

		if d.HasChange("manage_admin_password") {
			input.ManageAdminPassword = aws.Bool(d.Get("manage_admin_password").(bool))
		}

		_, err := conn.UpdateNamespace(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Namespace (%s): %s", d.Id(), err)
		}

		if _, err := waitNamespaceUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Namespace (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceNamespaceRead(ctx, d, meta)...)
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Namespace: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ConflictException](ctx, namespaceDeletedTimeout,
		func() (any, error) {
			return conn.DeleteNamespace(ctx, &redshiftserverless.DeleteNamespaceInput{
				NamespaceName: aws.String(d.Id()),
			})
		},
		// "ConflictException: There is an operation running on the namespace. Try deleting the namespace again later."
		"operation running")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

const (
	namespaceDeletedTimeout = 10 * time.Minute
	namespaceUpdatedTimeout = 10 * time.Minute
)

func findNamespaceByName(ctx context.Context, conn *redshiftserverless.Client, name string) (*awstypes.Namespace, error) {
	input := &redshiftserverless.GetNamespaceInput{
		NamespaceName: aws.String(name),
	}

	output, err := conn.GetNamespace(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Namespace, nil
}

func statusNamespace(ctx context.Context, conn *redshiftserverless.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findNamespaceByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitNamespaceDeleted(ctx context.Context, conn *redshiftserverless.Client, name string) (*awstypes.Namespace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NamespaceStatusDeleting),
		Target:  []string{},
		Refresh: statusNamespace(ctx, conn, name),
		Timeout: namespaceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Namespace); ok {
		return output, err
	}

	return nil, err
}

func waitNamespaceUpdated(ctx context.Context, conn *redshiftserverless.Client, name string) (*awstypes.Namespace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NamespaceStatusModifying),
		Target:  enum.Slice(awstypes.NamespaceStatusAvailable),
		Refresh: statusNamespace(ctx, conn, name),
		Timeout: namespaceUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Namespace); ok {
		return output, err
	}

	return nil, err
}

var (
	reIAMRole = regexache.MustCompile(`^\s*IamRole\((.*)\)\s*$`)
)

func flattenNamespaceIAMRoles(iamRoles []string) []string {
	var tfList []string

	for _, iamRole := range iamRoles {
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
