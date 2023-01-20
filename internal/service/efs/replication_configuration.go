package efs

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationConfigurationCreate,
		ReadWithoutTimeout:   resourceReplicationConfigurationRead,
		DeleteWithoutTimeout: resourceReplicationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							AtLeastOneOf: []string{"destination.0.availability_zone_name", "destination.0.region"},
						},
						"file_system_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"region": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidRegionName,
							AtLeastOneOf: []string{"destination.0.availability_zone_name", "destination.0.region"},
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"original_source_file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_file_system_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceReplicationConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	fsID := d.Get("source_file_system_id").(string)
	input := &efs.CreateReplicationConfigurationInput{
		SourceFileSystemId: aws.String(fsID),
	}

	if v, ok := d.GetOk("destination"); ok && len(v.([]interface{})) > 0 {
		input.Destinations = expandDestinationsToCreate(v.([]interface{}))
	}

	_, err := conn.CreateReplicationConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS Replication Configuration (%s): %s", fsID, err)
	}

	d.SetId(fsID)

	if _, err := waitReplicationConfigurationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS Replication Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReplicationConfigurationRead(ctx, d, meta)...)
}

func resourceReplicationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	replication, err := FindReplicationConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS Replication Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Replication Configuration (%s): %s", d.Id(), err)
	}

	destinations := flattenDestinations(replication.Destinations)

	// availability_zone_name and kms_key_id aren't returned from the AWS Read API.
	if v, ok := d.GetOk("destination"); ok && len(v.([]interface{})) > 0 {
		copy := func(i int, k string) {
			destinations[i].(map[string]interface{})[k] = v.([]interface{})[i].(map[string]interface{})[k]
		}
		// Assume 1 destination.
		copy(0, "availability_zone_name")
		copy(0, "kms_key_id")
	}

	d.Set("creation_time", aws.TimeValue(replication.CreationTime).String())
	if err := d.Set("destination", destinations); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting destination: %s", err)
	}
	d.Set("original_source_file_system_arn", replication.OriginalSourceFileSystemArn)
	d.Set("source_file_system_arn", replication.SourceFileSystemArn)
	d.Set("source_file_system_id", replication.SourceFileSystemId)
	d.Set("source_file_system_region", replication.SourceFileSystemRegion)

	return diags
}

func resourceReplicationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	// Deletion of the replication configuration must be done from the
	// Region in which the destination file system is located.
	destination := expandDestinationsToCreate(d.Get("destination").([]interface{}))[0]
	session, err := conns.NewSessionForRegion(&conn.Config, aws.StringValue(destination.Region), meta.(*conns.AWSClient).TerraformVersion)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AWS session: %s", err)
	}

	deleteConn := efs.New(session)

	log.Printf("[DEBUG] Deleting EFS Replication Configuration: %s", d.Id())
	_, err = deleteConn.DeleteReplicationConfigurationWithContext(ctx, &efs.DeleteReplicationConfigurationInput{
		SourceFileSystemId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound, efs.ErrCodeReplicationNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS Replication Configuration (%s): %s", d.Id(), err)
	}

	if _, err := waitReplicationConfigurationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS Replication Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandDestinationToCreate(tfMap map[string]interface{}) *efs.DestinationToCreate {
	if tfMap == nil {
		return nil
	}

	apiObject := &efs.DestinationToCreate{}

	if v, ok := tfMap["availability_zone_name"].(string); ok && v != "" {
		apiObject.AvailabilityZoneName = aws.String(v)
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["region"].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func expandDestinationsToCreate(tfList []interface{}) []*efs.DestinationToCreate {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*efs.DestinationToCreate

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDestinationToCreate(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDestination(apiObject *efs.Destination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FileSystemId; v != nil {
		tfMap["file_system_id"] = aws.StringValue(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.StringValue(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDestinations(apiObjects []*efs.Destination) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDestination(apiObject))
	}

	return tfList
}
