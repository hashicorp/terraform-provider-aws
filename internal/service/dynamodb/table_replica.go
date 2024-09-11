// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	resNameTableReplica = "Table Replica"
)

// @SDKResource("aws_dynamodb_table_replica", name="Table Replica")
// @Tags
func resourceTableReplica() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableReplicaCreate,
		ReadWithoutTimeout:   resourceTableReplicaRead,
		UpdateWithoutTimeout: resourceTableReplicaUpdate,
		DeleteWithoutTimeout: resourceTableReplicaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: { // direct to replica
				Type:     schema.TypeString,
				Computed: true,
			},
			// global_secondary_index read capacity override can be set but not return by aws atm either through main/replica nor directly
			"global_table_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrKMSKeyARN: { // through main table
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"point_in_time_recovery": { // direct to replica
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// read_capacity_override can be set but requires table write_capacity to be autoscaled which is not yet supported in the provider
			"table_class_override": { // through main table
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TableClass](),
			},
			names.AttrTags:    tftags.TagsSchema(),         // direct to replica
			names.AttrTagsAll: tftags.TagsSchemaComputed(), // direct to replica
		},
	}
}

func resourceTableReplicaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	replicaRegion := meta.(*conns.AWSClient).Region

	mainRegion, err := regionFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTableReplica, d.Get("global_table_arn").(string), err)
	}

	if mainRegion == replicaRegion {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTableReplica, d.Get("global_table_arn").(string), errors.New("replica cannot be in same region as main table"))
	}

	// now main table region
	optFn := func(o *dynamodb.Options) {
		o.Region = mainRegion
	}

	var replicaInput = &awstypes.CreateReplicationGroupMemberAction{}

	replicaInput.RegionName = aws.String(replicaRegion)

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		replicaInput.KMSMasterKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("table_class_override"); ok {
		replicaInput.TableClassOverride = awstypes.TableClass(v.(string))
	}

	tableName, err := tableNameFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTableReplica, d.Get("global_table_arn").(string), err)
	}

	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
		ReplicaUpdates: []awstypes.ReplicationGroupUpdate{{
			Create: replicaInput,
		}},
	}

	err = retry.RetryContext(ctx, max(replicaUpdateTimeout, d.Timeout(schema.TimeoutCreate)), func() *retry.RetryError {
		_, err := conn.UpdateTable(ctx, input, optFn)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
				return retry.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "simultaneously") {
				return retry.RetryableError(err)
			}
			if errs.IsA[*awstypes.ResourceInUseException](err) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateTable(ctx, input, optFn)
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTableReplica, d.Get("global_table_arn").(string), err)
	}

	if _, err := waitReplicaActive(ctx, conn, tableName, meta.(*conns.AWSClient).Region, d.Timeout(schema.TimeoutCreate), optFn); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameTableReplica, d.Get("global_table_arn").(string), err)
	}

	d.SetId(tableReplicaCreateResourceID(tableName, mainRegion))

	repARN, err := arnForNewRegion(d.Get("global_table_arn").(string), replicaRegion)
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameTableReplica, d.Id(), err)
	}

	d.Set(names.AttrARN, repARN)

	return append(diags, resourceTableReplicaUpdate(ctx, d, meta)...)
}

func resourceTableReplicaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// handled through main table (global table)
	// * global_secondary_index
	// * kms_key_arn
	// * read_capacity_override
	// * table_class_override
	diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	replicaRegion := meta.(*conns.AWSClient).Region

	tableName, mainRegion, err := tableReplicaParseResourceID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), err)
	}

	globalTableARN := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    mainRegion,
		Resource:  fmt.Sprintf("table/%s", tableName),
		Service:   "dynamodb",
	}.String()
	d.Set("global_table_arn", globalTableARN)

	if mainRegion == replicaRegion {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), errors.New("replica cannot be in same region as main table"))
	}

	// now main table region
	optFn := func(o *dynamodb.Options) {
		o.Region = mainRegion
	}

	table, err := findTableByName(ctx, conn, tableName, optFn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Dynamodb Table (%s) not found, removing replica from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), err)
	}

	replica := replicaForRegion(table.Replicas, replicaRegion)

	if !d.IsNewResource() && replica == nil {
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id())
		d.SetId("")
		return diags
	}

	if replica == nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), err)
	}

	dk, err := kms.FindDefaultKeyARNForService(ctx, meta.(*conns.AWSClient).KMSClient(ctx), "dynamodb", replicaRegion)
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), err)
	}

	if replica.KMSMasterKeyId == nil || aws.ToString(replica.KMSMasterKeyId) == dk {
		d.Set(names.AttrKMSKeyARN, nil)
	} else {
		d.Set(names.AttrKMSKeyARN, replica.KMSMasterKeyId)
	}

	if replica.ReplicaTableClassSummary != nil {
		d.Set("table_class_override", replica.ReplicaTableClassSummary.TableClass)
	} else {
		d.Set("table_class_override", nil)
	}

	return append(diags, resourceTableReplicaReadReplica(ctx, d, meta)...)
}

func resourceTableReplicaReadReplica(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// handled direct to replica
	// * arn
	// * point_in_time_recovery
	// * tags
	diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName, _, err := tableReplicaParseResourceID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), err)
	}

	table, err := findTableByName(ctx, conn, tableName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Dynamodb Table Replica (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), err)
	}

	d.Set(names.AttrARN, table.TableArn)

	pitrOut, err := conn.DescribeContinuousBackups(ctx, &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	})
	// When a Table is `ARCHIVED`, DescribeContinuousBackups returns `TableNotFoundException`
	if err != nil && !tfawserr.ErrCodeEquals(err, errCodeUnknownOperationException, errCodeTableNotFoundException) {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), fmt.Errorf("continuous backups: %w", err))
	}

	if table.SSEDescription != nil && table.SSEDescription.KMSMasterKeyArn != nil {
		d.Set(names.AttrKMSKeyARN, table.SSEDescription.KMSMasterKeyArn)
	}

	if pitrOut != nil && pitrOut.ContinuousBackupsDescription != nil && pitrOut.ContinuousBackupsDescription.PointInTimeRecoveryDescription != nil {
		d.Set("point_in_time_recovery", pitrOut.ContinuousBackupsDescription.PointInTimeRecoveryDescription.PointInTimeRecoveryStatus == awstypes.PointInTimeRecoveryStatusEnabled)
	} else {
		d.Set("point_in_time_recovery", false)
	}

	tags, err := listTags(ctx, conn, d.Get(names.AttrARN).(string))
	// When a Table is `ARCHIVED`, ListTags returns `ResourceNotFoundException`
	if err != nil && !(tfawserr.ErrMessageContains(err, errCodeUnknownOperationException, "Tagging is not currently supported in DynamoDB Local.") || tfresource.NotFound(err)) {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, resNameTableReplica, d.Id(), fmt.Errorf("tags: %w", err))
	}

	setTagsOut(ctx, Tags(tags))

	return diags
}

func resourceTableReplicaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// handled through main table (main)
	// * global_secondary_index
	// * kms_key_arn
	// * read_capacity_override
	// * table_class_override
	diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName, mainRegion, err := tableReplicaParseResourceID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTableReplica, d.Id(), err)
	}

	replicaRegion := meta.(*conns.AWSClient).Region

	if mainRegion == replicaRegion {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTableReplica, d.Id(), errors.New("replica cannot be in same region as main table"))
	}

	// now main table region
	optFn := func(o *dynamodb.Options) {
		o.Region = mainRegion
	}

	viaMainChanges := false
	viaMainInput := &awstypes.UpdateReplicationGroupMemberAction{
		RegionName: aws.String(replicaRegion),
	}

	if d.HasChange(names.AttrKMSKeyARN) && !d.IsNewResource() { // create ends with update and sets kms_key_arn causing change that is not
		dk, err := kms.FindDefaultKeyARNForService(ctx, meta.(*conns.AWSClient).KMSClient(ctx), "dynamodb", replicaRegion)
		if err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTableReplica, d.Id(), fmt.Errorf("region %s: %w", replicaRegion, err))
		}

		if d.Get(names.AttrKMSKeyARN).(string) != dk {
			viaMainChanges = true
			viaMainInput.KMSMasterKeyId = aws.String(d.Get(names.AttrKMSKeyARN).(string))
		}
	}

	if viaMainChanges {
		input := &dynamodb.UpdateTableInput{
			ReplicaUpdates: []awstypes.ReplicationGroupUpdate{{
				Update: viaMainInput,
			}},
			TableName: aws.String(tableName),
		}

		err := retry.RetryContext(ctx, max(replicaUpdateTimeout, d.Timeout(schema.TimeoutUpdate)), func() *retry.RetryError {
			_, err := conn.UpdateTable(ctx, input, optFn)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
					return retry.RetryableError(err)
				}
				if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
					return retry.RetryableError(err)
				}
				if errs.IsA[*awstypes.ResourceInUseException](err) {
					return retry.RetryableError(err)
				}

				return retry.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateTable(ctx, input, optFn)
		}

		if err != nil && !tfawserr.ErrMessageContains(err, errCodeValidationException, "no actions specified") {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTableReplica, d.Id(), err)
		}

		if _, err := waitReplicaActive(ctx, conn, tableName, replicaRegion, d.Timeout(schema.TimeoutUpdate), optFn); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTableReplica, d.Id(), err)
		}
	}

	// handled direct to replica
	// * point_in_time_recovery
	// * tags
	if d.HasChanges("point_in_time_recovery", names.AttrTagsAll) {
		if d.HasChange(names.AttrTagsAll) {
			o, n := d.GetChange(names.AttrTagsAll)
			if err := updateTags(ctx, conn, d.Get(names.AttrARN).(string), o, n); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTableReplica, d.Id(), err)
			}
		}

		if d.HasChange("point_in_time_recovery") {
			if err := updatePITR(ctx, conn, tableName, d.Get("point_in_time_recovery").(bool), replicaRegion, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameTableReplica, d.Id(), err)
			}
		}

		if _, err := waitReplicaActive(ctx, conn, tableName, replicaRegion, d.Timeout(schema.TimeoutUpdate), optFn); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameTableReplica, d.Id(), err)
		}
	}

	return append(diags, resourceTableReplicaRead(ctx, d, meta)...)
}

func resourceTableReplicaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName, mainRegion, err := tableReplicaParseResourceID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionDeleting, resNameTableReplica, d.Id(), err)
	}

	replicaRegion := meta.(*conns.AWSClient).Region

	// now main table region.
	optFn := func(o *dynamodb.Options) {
		o.Region = mainRegion
	}

	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
		ReplicaUpdates: []awstypes.ReplicationGroupUpdate{
			{
				Delete: &awstypes.DeleteReplicationGroupMemberAction{
					RegionName: aws.String(replicaRegion),
				},
			},
		},
	}

	err = retry.RetryContext(ctx, updateTableTimeout, func() *retry.RetryError {
		_, err := conn.UpdateTable(ctx, input, optFn)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeThrottlingException) {
				return retry.RetryableError(err)
			}
			if errs.IsAErrorMessageContains[*awstypes.LimitExceededException](err, "can be created, updated, or deleted simultaneously") {
				return retry.RetryableError(err)
			}
			if errs.IsA[*awstypes.ResourceInUseException](err) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateTable(ctx, input, optFn)
	}

	if tfawserr.ErrMessageContains(err, errCodeValidationException, "Replica specified in the Replica Update or Replica Delete action of the request was not found") {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionDeleting, resNameTableReplica, d.Id(), err)
	}

	if _, err := waitReplicaDeleted(ctx, conn, tableName, replicaRegion, d.Timeout(schema.TimeoutDelete), optFn); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForDeletion, resNameTableReplica, d.Id(), err)
	}

	return diags
}

const tableReplicaResourceIDSeparator = ":"

func tableReplicaCreateResourceID(tableName, mainRegion string) string {
	parts := []string{tableName, mainRegion}
	id := strings.Join(parts, tableReplicaResourceIDSeparator)

	return id
}

func tableReplicaParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected table-name:main-table-region", id)
}

func replicaForRegion(replicas []awstypes.ReplicaDescription, region string) *awstypes.ReplicaDescription {
	for _, replica := range replicas {
		if aws.ToString(replica.RegionName) == region {
			return &replica
		}
	}

	return nil
}
