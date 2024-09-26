// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_efs_replication_configuration", name="Replication Configuration")
func resourceReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationConfigurationCreate,
		ReadWithoutTimeout:   resourceReplicationConfigurationRead,
		DeleteWithoutTimeout: resourceReplicationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDestination: {
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
						names.AttrFileSystemID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrRegion: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidRegionName,
							AtLeastOneOf: []string{"destination.0.availability_zone_name", "destination.0.region"},
						},
						names.AttrStatus: {
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
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	fsID := d.Get("source_file_system_id").(string)
	input := &efs.CreateReplicationConfigurationInput{
		SourceFileSystemId: aws.String(fsID),
	}

	if v, ok := d.GetOk(names.AttrDestination); ok && len(v.([]interface{})) > 0 {
		input.Destinations = expandDestinationsToCreate(v.([]interface{}))
	}

	_, err := conn.CreateReplicationConfiguration(ctx, input)

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
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	replication, err := findReplicationConfigurationByID(ctx, conn, d.Id())

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
	if v, ok := d.GetOk(names.AttrDestination); ok && len(v.([]interface{})) > 0 {
		copy := func(i int, k string) {
			destinations[i].(map[string]interface{})[k] = v.([]interface{})[i].(map[string]interface{})[k]
		}
		// Assume 1 destination.
		copy(0, "availability_zone_name")
		copy(0, names.AttrKMSKeyID)
	}

	d.Set(names.AttrCreationTime, aws.ToTime(replication.CreationTime).String())
	if err := d.Set(names.AttrDestination, destinations); err != nil {
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
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	// Deletion of the replication configuration must be done from the Region in which the destination file system is located.
	destination := expandDestinationsToCreate(d.Get(names.AttrDestination).([]interface{}))[0]
	optFn := func(o *efs.Options) {
		o.Region = aws.ToString(destination.Region)
	}

	log.Printf("[DEBUG] Deleting EFS Replication Configuration: %s", d.Id())
	if err := deleteReplicationConfiguration(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete), optFn); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// Delete also in the source Region.
	if err := deleteReplicationConfiguration(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func deleteReplicationConfiguration(ctx context.Context, conn *efs.Client, fsID string, timeout time.Duration, optFns ...func(*efs.Options)) error {
	_, err := conn.DeleteReplicationConfiguration(ctx, &efs.DeleteReplicationConfigurationInput{
		SourceFileSystemId: aws.String(fsID),
	}, optFns...)

	if errs.IsA[*awstypes.FileSystemNotFound](err) || errs.IsA[*awstypes.ReplicationNotFound](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EFS Replication Configuration (%s): %w", fsID, err)
	}

	if _, err := waitReplicationConfigurationDeleted(ctx, conn, fsID, timeout, optFns...); err != nil {
		return fmt.Errorf("waiting for EFS Replication Configuration (%s) delete: %w", fsID, err)
	}

	return nil
}

func findReplicationConfiguration(ctx context.Context, conn *efs.Client, input *efs.DescribeReplicationConfigurationsInput, filter tfslices.Predicate[*awstypes.ReplicationConfigurationDescription], optFns ...func(*efs.Options)) (*awstypes.ReplicationConfigurationDescription, error) {
	output, err := findReplicationConfigurations(ctx, conn, input, filter, optFns...)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReplicationConfigurations(ctx context.Context, conn *efs.Client, input *efs.DescribeReplicationConfigurationsInput, filter tfslices.Predicate[*awstypes.ReplicationConfigurationDescription], optFns ...func(*efs.Options)) ([]awstypes.ReplicationConfigurationDescription, error) {
	var output []awstypes.ReplicationConfigurationDescription

	pages := efs.NewDescribeReplicationConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if errs.IsA[*awstypes.FileSystemNotFound](err) || errs.IsA[*awstypes.ReplicationNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Replications {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findReplicationConfigurationByID(ctx context.Context, conn *efs.Client, id string, optFns ...func(*efs.Options)) (*awstypes.ReplicationConfigurationDescription, error) {
	input := &efs.DescribeReplicationConfigurationsInput{
		FileSystemId: aws.String(id),
	}

	output, err := findReplicationConfiguration(ctx, conn, input, tfslices.PredicateTrue[*awstypes.ReplicationConfigurationDescription](), optFns...)

	if err != nil {
		return nil, err
	}

	if len(output.Destinations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusReplicationConfiguration(ctx context.Context, conn *efs.Client, id string, optFns ...func(*efs.Options)) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReplicationConfigurationByID(ctx, conn, id, optFns...)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Destinations[0].Status), nil
	}
}

func waitReplicationConfigurationCreated(ctx context.Context, conn *efs.Client, id string, timeout time.Duration, optFns ...func(*efs.Options)) (*awstypes.ReplicationConfigurationDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ReplicationStatusEnabling),
		Target:  enum.Slice(awstypes.ReplicationStatusEnabled),
		Refresh: statusReplicationConfiguration(ctx, conn, id, optFns...),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationConfigurationDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationConfigurationDeleted(ctx context.Context, conn *efs.Client, id string, timeout time.Duration, optFns ...func(*efs.Options)) (*awstypes.ReplicationConfigurationDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ReplicationStatusDeleting),
		Target:                    []string{},
		Refresh:                   statusReplicationConfiguration(ctx, conn, id, optFns...),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationConfigurationDescription); ok {
		return output, err
	}

	return nil, err
}

func expandDestinationToCreate(tfMap map[string]interface{}) *awstypes.DestinationToCreate {
	apiObject := &awstypes.DestinationToCreate{}

	if v, ok := tfMap["availability_zone_name"].(string); ok && v != "" {
		apiObject.AvailabilityZoneName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrFileSystemID].(string); ok && v != "" {
		apiObject.FileSystemId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func expandDestinationsToCreate(tfList []interface{}) []awstypes.DestinationToCreate {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DestinationToCreate

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandDestinationToCreate(tfMap)

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenDestination(apiObject awstypes.Destination) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.FileSystemId; v != nil {
		tfMap[names.AttrFileSystemID] = aws.ToString(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	tfMap[names.AttrStatus] = string(apiObject.Status)

	return tfMap
}

func flattenDestinations(apiObjects []awstypes.Destination) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDestination(apiObject))
	}

	return tfList
}
