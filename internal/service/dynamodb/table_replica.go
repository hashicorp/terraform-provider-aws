package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameTableReplica = "Table Replica"
)

func ResourceTableReplica() *schema.Resource {
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
				ValidateFunc: verify.ValidARN,
				ForceNew:     true,
			},
			"point_in_time_recovery": { // direct to replica
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// read_capacity_override can be set but requires table write_capacity to be autoscaled which is not yet supported in the provider
			"table_class_override": { // through main table
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(dynamodb.TableClass_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),         // direct to replica
			names.AttrTagsAll: tftags.TagsSchemaComputed(), // direct to replica
		},
	}
}

func resourceTableReplicaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn()

	replicaRegion := aws.StringValue(conn.Config.Region)

	mainRegion, err := RegionFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionCreating, ResNameTableReplica, d.Get("global_table_arn").(string), err)
	}

	if mainRegion == aws.StringValue(conn.Config.Region) {
		return create.DiagError(names.DynamoDB, create.ErrActionCreating, ResNameTableReplica, d.Get("global_table_arn").(string), errors.New("replica cannot be in same region as main table"))
	}

	session, err := conns.NewSessionForRegion(&conn.Config, mainRegion, meta.(*conns.AWSClient).TerraformVersion)
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionCreating, ResNameTableReplica, d.Get("global_table_arn").(string), fmt.Errorf("region %s: %w", mainRegion, err))
	}

	conn = dynamodb.New(session) // now main table region

	var replicaInput = &dynamodb.CreateReplicationGroupMemberAction{}

	replicaInput.RegionName = aws.String(replicaRegion)

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		replicaInput.KMSMasterKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("table_class_override"); ok {
		replicaInput.TableClassOverride = aws.String(v.(string))
	}

	tableName, err := TableNameFromARN(d.Get("global_table_arn").(string))
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionCreating, ResNameTableReplica, d.Get("global_table_arn").(string), err)
	}

	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
		ReplicaUpdates: []*dynamodb.ReplicationGroupUpdate{{
			Create: replicaInput,
		}},
	}

	err = resource.RetryContext(ctx, maxDuration(replicaUpdateTimeout, d.Timeout(schema.TimeoutCreate)), func() *resource.RetryError {
		_, err := conn.UpdateTableWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, "ThrottlingException") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceInUseException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateTableWithContext(ctx, input)
	}

	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionCreating, ResNameTableReplica, d.Get("global_table_arn").(string), err)
	}

	if err := waitReplicaActive(ctx, conn, tableName, meta.(*conns.AWSClient).Region, d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionWaitingForCreation, ResNameTableReplica, d.Get("global_table_arn").(string), err)
	}

	d.SetId(tableReplicaID(tableName, mainRegion))

	repARN, err := ARNForNewRegion(d.Get("global_table_arn").(string), replicaRegion)
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionCreating, ResNameTableReplica, d.Id(), err)
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

	conn := meta.(*conns.AWSClient).DynamoDBConn()

	replicaRegion := aws.StringValue(conn.Config.Region)

	tableName, mainRegion, err := TableReplicaParseID(d.Id())
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), err)
	}

	globalTableARN := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    mainRegion,
		Resource:  fmt.Sprintf("table/%s", tableName),
		Service:   dynamodb.EndpointsID,
	}.String()

	d.Set("global_table_arn", globalTableARN)

	if mainRegion == replicaRegion {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), errors.New("replica cannot be in same region as main table"))
	}

	session, err := conns.NewSessionForRegion(&conn.Config, mainRegion, meta.(*conns.AWSClient).TerraformVersion)
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), fmt.Errorf("region %s: %w", mainRegion, err))
	}

	conn = dynamodb.New(session) // now main table region

	result, err := conn.DescribeTableWithContext(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Dynamodb Table (%s) not found, removing replica from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), err)
	}

	if result == nil || result.Table == nil {
		if d.IsNewResource() {
			return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), errors.New("empty output after creation"))
		}
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id())
		d.SetId("")
		return diags
	}

	replica, err := FilterReplicasByRegion(result.Table.Replicas, replicaRegion)
	if !d.IsNewResource() && err != nil && err.Error() == "no replicas found" {
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), err)
	}

	dk, err := kms.FindDefaultKey(ctx, "dynamodb", replicaRegion, meta)
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), err)
	}

	if replica.KMSMasterKeyId == nil || aws.StringValue(replica.KMSMasterKeyId) == dk {
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

	conn := meta.(*conns.AWSClient).DynamoDBConn()

	tableName, _, err := TableReplicaParseID(d.Id())
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), err)
	}

	result, err := conn.DescribeTableWithContext(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Dynamodb Table Replica (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), err)
	}

	if result == nil || result.Table == nil {
		if d.IsNewResource() {
			return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), errors.New("empty output after creation"))
		}
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id())
		d.SetId("")
		return diags
	}

	d.Set(names.AttrARN, result.Table.TableArn)

	pitrOut, err := conn.DescribeContinuousBackupsWithContext(ctx, &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	})
	// When a Table is `ARCHIVED`, DescribeContinuousBackups returns `TableNotFoundException`
	if err != nil && !tfawserr.ErrCodeEquals(err, "UnknownOperationException", dynamodb.ErrCodeTableNotFoundException) {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), fmt.Errorf("continuous backups: %w", err))
	}

	if pitrOut != nil && pitrOut.ContinuousBackupsDescription != nil && pitrOut.ContinuousBackupsDescription.PointInTimeRecoveryDescription != nil {
		d.Set("point_in_time_recovery", aws.StringValue(pitrOut.ContinuousBackupsDescription.PointInTimeRecoveryDescription.PointInTimeRecoveryStatus) == dynamodb.PointInTimeRecoveryStatusEnabled)
	} else {
		d.Set("point_in_time_recovery", false)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	tags, err := ListTags(ctx, conn, d.Get(names.AttrARN).(string))
	// When a Table is `ARCHIVED`, ListTags returns `ResourceNotFoundException`
	if err != nil && !(tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") || tfresource.NotFound(err)) {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, ResNameTableReplica, d.Id(), fmt.Errorf("tags: %w", err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagSettingError(names.DynamoDB, ResNameTableReplica, d.Id(), names.AttrTags, err)
	}

	if err := d.Set(names.AttrTagsAll, tags.Map()); err != nil {
		return create.DiagSettingError(names.DynamoDB, ResNameTableReplica, d.Id(), names.AttrTagsAll, err)
	}

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

	repConn := meta.(*conns.AWSClient).DynamoDBConn()

	tableName, mainRegion, err := TableReplicaParseID(d.Id())
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionUpdating, ResNameTableReplica, d.Id(), err)
	}

	replicaRegion := aws.StringValue(repConn.Config.Region)

	if mainRegion == replicaRegion {
		return create.DiagError(names.DynamoDB, create.ErrActionUpdating, ResNameTableReplica, d.Id(), errors.New("replica cannot be in same region as main table"))
	}

	session, err := conns.NewSessionForRegion(&repConn.Config, mainRegion, meta.(*conns.AWSClient).TerraformVersion)
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionUpdating, ResNameTableReplica, d.Id(), fmt.Errorf("region %s: %w", mainRegion, err))
	}

	tabConn := dynamodb.New(session) // now main table region

	viaMainChanges := false
	viaMainInput := &dynamodb.UpdateReplicationGroupMemberAction{
		RegionName: aws.String(replicaRegion),
	}

	if d.HasChange(names.AttrKMSKeyARN) && !d.IsNewResource() { // create ends with update and sets kms_key_arn causing change that is not
		dk, err := kms.FindDefaultKey(ctx, "dynamodb", replicaRegion, meta)
		if err != nil {
			return create.DiagError(names.DynamoDB, create.ErrActionUpdating, ResNameTableReplica, d.Id(), fmt.Errorf("region %s: %w", replicaRegion, err))
		}

		if d.Get(names.AttrKMSKeyARN).(string) != dk {
			viaMainChanges = true
			viaMainInput.KMSMasterKeyId = aws.String(d.Get(names.AttrKMSKeyARN).(string))
		}
	}

	if viaMainChanges {
		input := &dynamodb.UpdateTableInput{
			ReplicaUpdates: []*dynamodb.ReplicationGroupUpdate{{
				Update: viaMainInput,
			}},
			TableName: aws.String(tableName),
		}

		err := resource.RetryContext(ctx, maxDuration(replicaUpdateTimeout, d.Timeout(schema.TimeoutUpdate)), func() *resource.RetryError {
			_, err := tabConn.UpdateTableWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, "ThrottlingException") {
					return resource.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "can be created, updated, or deleted simultaneously") {
					return resource.RetryableError(err)
				}
				if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceInUseException) {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = tabConn.UpdateTableWithContext(ctx, input)
		}

		if err != nil && !tfawserr.ErrMessageContains(err, "ValidationException", "no actions specified") {
			return create.DiagError(names.DynamoDB, create.ErrActionUpdating, ResNameTableReplica, d.Id(), err)
		}

		if err := waitReplicaActive(ctx, tabConn, tableName, replicaRegion, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.DynamoDB, create.ErrActionWaitingForUpdate, ResNameTableReplica, d.Id(), err)
		}
	}

	// handled direct to replica
	// * point_in_time_recovery
	// * tags
	if d.HasChanges("point_in_time_recovery", names.AttrTagsAll) {
		if d.HasChange(names.AttrTagsAll) {
			o, n := d.GetChange(names.AttrTagsAll)
			if err := UpdateTags(ctx, repConn, d.Get(names.AttrARN).(string), o, n); err != nil {
				return create.DiagError(names.DynamoDB, create.ErrActionUpdating, ResNameTableReplica, d.Id(), err)
			}
		}

		if d.HasChange("point_in_time_recovery") {
			if err := updatePITR(ctx, repConn, tableName, d.Get("point_in_time_recovery").(bool), replicaRegion, meta.(*conns.AWSClient).TerraformVersion, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return create.DiagError(names.DynamoDB, create.ErrActionUpdating, ResNameTableReplica, d.Id(), err)
			}
		}

		if err := waitReplicaActive(ctx, tabConn, tableName, replicaRegion, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.DynamoDB, create.ErrActionWaitingForUpdate, ResNameTableReplica, d.Id(), err)
		}
	}

	return append(diags, resourceTableReplicaRead(ctx, d, meta)...)
}

func resourceTableReplicaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[DEBUG] DynamoDB delete Table Replica: %s", d.Id())

	tableName, mainRegion, err := TableReplicaParseID(d.Id())
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionDeleting, ResNameTableReplica, d.Id(), err)
	}

	conn := meta.(*conns.AWSClient).DynamoDBConn()

	replicaRegion := aws.StringValue(conn.Config.Region)

	session, err := conns.NewSessionForRegion(&conn.Config, mainRegion, meta.(*conns.AWSClient).TerraformVersion)
	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionDeleting, ResNameTableReplica, d.Id(), fmt.Errorf("region %s: %w", mainRegion, err))
	}

	conn = dynamodb.New(session) // now main table region

	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
		ReplicaUpdates: []*dynamodb.ReplicationGroupUpdate{
			{
				Delete: &dynamodb.DeleteReplicationGroupMemberAction{
					RegionName: aws.String(replicaRegion),
				},
			},
		},
	}

	err = resource.RetryContext(ctx, updateTableTimeout, func() *resource.RetryError {
		_, err := conn.UpdateTableWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, "ThrottlingException") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeLimitExceededException, "can be created, updated, or deleted simultaneously") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceInUseException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateTableWithContext(ctx, input)
	}

	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionDeleting, ResNameTableReplica, d.Id(), err)
	}

	if err := waitReplicaDeleted(ctx, conn, tableName, replicaRegion, d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionWaitingForDeletion, ResNameTableReplica, d.Id(), err)
	}

	return diags
}

func TableReplicaParseID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected table-name:main-table-region", id)
}

func tableReplicaID(tableName, mainRegion string) string {
	return fmt.Sprintf("%s:%s", tableName, mainRegion)
}

func FilterReplicasByRegion(replicas []*dynamodb.ReplicaDescription, region string) (*dynamodb.ReplicaDescription, error) {
	if len(replicas) == 0 {
		return nil, errors.New("no replicas found")
	}

	for _, replica := range replicas {
		if aws.StringValue(replica.RegionName) == region {
			return replica, nil
		}
	}

	return nil, errors.New("replica not found")
}
