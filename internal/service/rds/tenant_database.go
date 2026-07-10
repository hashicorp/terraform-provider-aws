// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_tenant_database", name="Tenant Database")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceTenantDatabase() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTenantDatabaseCreate,
		ReadWithoutTimeout:   resourceTenantDatabaseRead,
		UpdateWithoutTimeout: resourceTenantDatabaseUpdate,
		DeleteWithoutTimeout: resourceTenantDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"character_set_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"db_instance_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"manage_master_user_password": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"master_password"},
			},
			"master_password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"manage_master_user_password"},
			},
			"master_user_secret": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"master_user_secret_kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"nchar_character_set_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tenant_database_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_db_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrUsername: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTenantDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	name := d.Get("tenant_db_name").(string)
	input := rds.CreateTenantDatabaseInput{
		DBInstanceIdentifier: aws.String(d.Get("db_instance_identifier").(string)),
		MasterUsername:       aws.String(d.Get(names.AttrUsername).(string)),
		Tags:                 getTagsIn(ctx),
		TenantDBName:         aws.String(name),
	}

	if v, ok := d.GetOk("character_set_name"); ok {
		input.CharacterSetName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("manage_master_user_password"); ok {
		input.ManageMasterUserPassword = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("master_password"); ok {
		input.MasterUserPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
		input.MasterUserSecretKmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("nchar_character_set_name"); ok {
		input.NcharCharacterSetName = aws.String(v.(string))
	}

	output, err := tfresource.RetryWhenIsA[*rds.CreateTenantDatabaseOutput, *types.InvalidDBInstanceStateFault](ctx, d.Timeout(schema.TimeoutCreate), func(ctx context.Context) (*rds.CreateTenantDatabaseOutput, error) {
		return conn.CreateTenantDatabase(ctx, &input)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Tenant Database (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TenantDatabase.TenantDatabaseResourceId))

	if _, err := waitTenantDatabaseAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Tenant Database (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTenantDatabaseRead(ctx, d, meta)...)
}

func resourceTenantDatabaseRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	tenantDB, err := findTenantDatabaseByID(ctx, conn, d.Id())
	if !d.IsNewResource() && retry.NotFound(err) {
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Tenant Database (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, tenantDB.TenantDatabaseARN)
	d.Set("character_set_name", tenantDB.CharacterSetName)
	d.Set("db_instance_identifier", tenantDB.DBInstanceIdentifier)
	// manage_master_user_password is a virtual attribute not returned by the API.
	// Infer it from the presence of MasterUserSecret so import works correctly.
	if ms := tenantDB.MasterUserSecret; ms != nil {
		if err := d.Set("master_user_secret", []any{flattenManagedMasterUserSecret(ms)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_user_secret: %s", err)
		}
		d.Set("manage_master_user_password", true)
	} else {
		d.Set("master_user_secret", nil)
	}
	d.Set("nchar_character_set_name", tenantDB.NcharCharacterSetName)
	d.Set(names.AttrStatus, tenantDB.Status)
	d.Set("tenant_db_name", tenantDB.TenantDBName)
	d.Set("tenant_database_resource_id", tenantDB.TenantDatabaseResourceId)
	d.Set(names.AttrUsername, tenantDB.MasterUsername)

	setTagsOut(ctx, tenantDB.TagList)

	return diags
}

func resourceTenantDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChanges("tenant_db_name", "master_password", "manage_master_user_password", "master_user_secret_kms_key_id") {
		tenantDB, err := findTenantDatabaseByID(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading RDS Tenant Database (%s): %s", d.Id(), err)
		}

		input := rds.ModifyTenantDatabaseInput{
			DBInstanceIdentifier: tenantDB.DBInstanceIdentifier,
			TenantDBName:         tenantDB.TenantDBName,
		}

		if d.HasChange("tenant_db_name") {
			input.NewTenantDBName = aws.String(d.Get("tenant_db_name").(string))
		}

		if d.HasChange("master_password") {
			input.MasterUserPassword = aws.String(d.Get("master_password").(string))
		}

		if d.HasChange("manage_master_user_password") {
			input.ManageMasterUserPassword = aws.Bool(d.Get("manage_master_user_password").(bool))
		}

		if d.HasChange("master_user_secret_kms_key_id") {
			input.MasterUserSecretKmsKeyId = aws.String(d.Get("master_user_secret_kms_key_id").(string))
		}

		_, err = conn.ModifyTenantDatabase(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS Tenant Database (%s): %s", d.Id(), err)
		}

		if _, err := waitTenantDatabaseAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Tenant Database (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTenantDatabaseRead(ctx, d, meta)...)
}

func resourceTenantDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	tenantDB, err := findTenantDatabaseByID(ctx, conn, d.Id())
	if retry.NotFound(err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Tenant Database (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting RDS Tenant Database: %s", d.Id())
	input := rds.DeleteTenantDatabaseInput{
		DBInstanceIdentifier: tenantDB.DBInstanceIdentifier,
		TenantDBName:         tenantDB.TenantDBName,
		SkipFinalSnapshot:    aws.Bool(true),
	}

	_, err = tfresource.RetryWhenIsA[*rds.DeleteTenantDatabaseOutput, *types.InvalidDBInstanceStateFault](ctx, d.Timeout(schema.TimeoutDelete), func(ctx context.Context) (*rds.DeleteTenantDatabaseOutput, error) {
		return conn.DeleteTenantDatabase(ctx, &input)
	})
	if errs.IsA[*types.DBInstanceNotFoundFault](err) || errs.IsA[*types.TenantDatabaseNotFoundFault](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Tenant Database (%s): %s", d.Id(), err)
	}

	if _, err := waitTenantDatabaseDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Tenant Database (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findTenantDatabaseByID(ctx context.Context, conn *rds.Client, id string) (*types.TenantDatabase, error) {
	input := &rds.DescribeTenantDatabasesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tenant-database-resource-id"),
				Values: []string{id},
			},
		},
	}

	return findTenantDatabase(ctx, conn, input, tfslices.PredicateTrue[*types.TenantDatabase]())
}

func findTenantDatabase(ctx context.Context, conn *rds.Client, input *rds.DescribeTenantDatabasesInput, filter tfslices.Predicate[*types.TenantDatabase]) (*types.TenantDatabase, error) {
	output, err := findTenantDatabases(ctx, conn, input, filter)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTenantDatabases(ctx context.Context, conn *rds.Client, input *rds.DescribeTenantDatabasesInput, filter tfslices.Predicate[*types.TenantDatabase]) ([]types.TenantDatabase, error) {
	var output []types.TenantDatabase

	pages := rds.NewDescribeTenantDatabasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBInstanceNotFoundFault](err) || errs.IsA[*types.TenantDatabaseNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.TenantDatabases {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusTenantDatabase(conn *rds.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTenantDatabaseByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitTenantDatabaseAvailable(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.TenantDatabase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			instanceStatusCreating,
			instanceStatusModifying,
			instanceStatusRenaming,
			instanceStatusResettingMasterCredentials,
		},
		Target:                    []string{instanceStatusAvailable},
		Refresh:                   statusTenantDatabase(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
		Delay:                     10 * time.Second,
		MinTimeout:                3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.TenantDatabase); ok {
		return output, err
	}

	return nil, err
}

func waitTenantDatabaseDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.TenantDatabase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			instanceStatusAvailable,
			instanceStatusDeleting,
			instanceStatusModifying,
		},
		Target:  []string{},
		Refresh: statusTenantDatabase(conn, id),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.TenantDatabase); ok {
		return output, err
	}

	return nil, err
}
