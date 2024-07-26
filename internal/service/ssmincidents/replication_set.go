// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameReplicationSet = "Replication Set"
)

// @SDKResource("aws_ssmincidents_replication_set", name="Replication Set")
// @Tags(identifierAttribute="id")
func ResourceReplicationSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationSetCreate,
		ReadWithoutTimeout:   resourceReplicationSetRead,
		UpdateWithoutTimeout: resourceReplicationSetUpdate,
		DeleteWithoutTimeout: resourceReplicationSetDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRegion: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrKMSKeyARN: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "DefaultKey",
							ValidateDiagFunc: validateNonAliasARN,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			// all other computed fields in alphabetic order
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protected": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"last_modified_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceReplicationSetImport,
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReplicationSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	input := &ssmincidents.CreateReplicationSetInput{
		Regions: expandRegions(d.Get(names.AttrRegion).(*schema.Set).List()),
		Tags:    getTagsIn(ctx),
	}

	createReplicationSetOutput, err := client.CreateReplicationSet(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionCreating, ResNameReplicationSet, "", err)
	}

	if createReplicationSetOutput == nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionCreating, ResNameReplicationSet, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(createReplicationSetOutput.Arn))

	getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(d.Id()),
	}

	if err := ssmincidents.NewWaitForReplicationSetActiveWaiter(client).Wait(ctx, getReplicationSetInput, d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionWaitingForCreation, ResNameReplicationSet, d.Id(), err)
	}

	return append(diags, resourceReplicationSetRead(ctx, d, meta)...)
}

func resourceReplicationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	replicationSet, err := FindReplicationSetByID(ctx, client, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSMIncidents ReplicationSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, d.Id(), err)
	}

	d.Set(names.AttrARN, replicationSet.Arn)
	d.Set("created_by", replicationSet.CreatedBy)
	d.Set("deletion_protected", replicationSet.DeletionProtected)
	d.Set("last_modified_by", replicationSet.LastModifiedBy)
	if err := d.Set(names.AttrRegion, flattenRegions(replicationSet.RegionMap)); err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, d.Id(), err)
	}
	d.Set(names.AttrStatus, replicationSet.Status)

	return diags
}

func resourceReplicationSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	if d.HasChanges(names.AttrRegion) {
		input := &ssmincidents.UpdateReplicationSetInput{
			Arn: aws.String(d.Id()),
		}

		if err := updateRegionsInput(d, input); err != nil {
			return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}

		log.Printf("[DEBUG] Updating SSMIncidents ReplicationSet (%s): %#v", d.Id(), input)
		_, err := client.UpdateReplicationSet(ctx, input)
		if err != nil {
			return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}

		getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
			Arn: aws.String(d.Id()),
		}

		if err := ssmincidents.NewWaitForReplicationSetActiveWaiter(client).Wait(ctx, getReplicationSetInput, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionWaitingForUpdate, ResNameReplicationSet, d.Id(), err)
		}
	}

	return append(diags, resourceReplicationSetRead(ctx, d, meta)...)
}

func resourceReplicationSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	log.Printf("[INFO] Deleting SSMIncidents ReplicationSet: %s", d.Id())
	_, err := client.DeleteReplicationSet(ctx, &ssmincidents.DeleteReplicationSetInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		var notFoundError *types.ResourceNotFoundException
		if errors.As(err, &notFoundError) {
			return diags
		}

		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionDeleting, ResNameReplicationSet, d.Id(), err)
	}

	getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(d.Id()),
	}

	if err := ssmincidents.NewWaitForReplicationSetDeletedWaiter(client).Wait(ctx, getReplicationSetInput, d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionWaitingForDeletion, ResNameReplicationSet, d.Id(), err)
	}

	return diags
}

// converts a list of regions to a map which maps region name to kms key arn
func regionListToRegionMap(list []interface{}) map[string]string {
	regionMap := make(map[string]string)
	for _, region := range list {
		regionData := region.(map[string]interface{})
		regionName := regionData[names.AttrName].(string)
		regionMap[regionName] = regionData[names.AttrKMSKeyARN].(string)
	}

	return regionMap
}

// updates UpdateReplicationSetInput to include any required actions
// invalid updates return errors from AWS Api
func updateRegionsInput(d *schema.ResourceData, input *ssmincidents.UpdateReplicationSetInput) error {
	old, new := d.GetChange(names.AttrRegion)
	oldRegions := regionListToRegionMap(old.(*schema.Set).List())
	newRegions := regionListToRegionMap(new.(*schema.Set).List())

	for region, oldcmk := range oldRegions {
		if newcmk, ok := newRegions[region]; !ok {
			// this region has been destroyed

			action := &types.UpdateReplicationSetActionMemberDeleteRegionAction{
				Value: types.DeleteRegionAction{
					RegionName: aws.String(region),
				},
			}

			input.Actions = append(input.Actions, action)
		} else {
			if oldcmk != newcmk {
				return fmt.Errorf("error: Incident Manager does not support updating encryption on a Replication Set's region. To do this, remove the region, and then re-create it with the new key")
			}
		}
	}

	for region, newcmk := range newRegions {
		if _, ok := oldRegions[region]; !ok {
			// this region is newly created

			action := &types.UpdateReplicationSetActionMemberAddRegionAction{
				Value: types.AddRegionAction{
					RegionName: aws.String(region),
				},
			}

			if newcmk != "DefaultKey" {
				action.Value.SseKmsKeyId = aws.String(newcmk)
			}

			input.Actions = append(input.Actions, action)
		}
	}

	return nil
}

func resourceReplicationSetImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	arn, err := getReplicationSetARN(ctx, client)

	if err != nil {
		return nil, err
	}

	d.SetId(arn)

	return []*schema.ResourceData{d}, nil
}

func validateNonAliasARN(value interface{}, path cty.Path) diag.Diagnostics {
	parsedARN, err := arn.Parse(value.(string))
	var diags diag.Diagnostics

	if err != nil || strings.HasPrefix(parsedARN.Resource, "alias/") {
		diag := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid kms_key_arn",
			Detail:   "Failed to parse key ARN. ARN must be a valid key ARN, not a key ID, or alias ARN",
		}
		diags = append(diags, diag)
	}

	return diags
}
