package ssmincidents

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
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

func ResourceReplicationSet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReplicationSetCreate,
		ReadContext:   resourceReplicationSetRead,
		UpdateContext: resourceReplicationSetUpdate,
		DeleteContext: resourceReplicationSetDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"kms_key_arn": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "DefaultKey",
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_update_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			// all other computed fields in alphabetic order
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
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
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
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

func resourceReplicationSetCreate(context context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	input := &ssmincidents.CreateReplicationSetInput{
		Regions: ExpandRegions(resourceData.Get("region").(*schema.Set).List()),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(resourceData.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAWS().Map()
	}

	createReplicationSetOutput, err := client.CreateReplicationSet(context, input)
	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionCreating, ResNameReplicationSet, "", err)
	}

	if createReplicationSetOutput == nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionCreating, ResNameReplicationSet, "", errors.New("empty output"))
	}

	resourceData.SetId(aws.ToString(createReplicationSetOutput.Arn))

	getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(resourceData.Id()),
	}

	if err := ssmincidents.NewWaitForReplicationSetActiveWaiter(client).Wait(context, getReplicationSetInput, resourceData.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForCreation, ResNameReplicationSet, resourceData.Id(), err)
	}

	return resourceReplicationSetRead(context, resourceData, meta)
}

func resourceReplicationSetRead(context context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	replicationSet, err := FindReplicationSetByID(context, client, resourceData.Id())

	if !resourceData.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSMIncidents ReplicationSet (%s) not found, removing from state", resourceData.Id())
		resourceData.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, resourceData.Id(), err)
	}

	resourceData.Set("arn", replicationSet.Arn)
	resourceData.Set("created_by", replicationSet.CreatedBy)
	resourceData.Set("created_time", replicationSet.CreatedTime.String())
	resourceData.Set("deletion_protected", replicationSet.DeletionProtected)
	resourceData.Set("last_modified_by", replicationSet.LastModifiedBy)
	resourceData.Set("last_modified_time", replicationSet.LastModifiedTime.String())
	resourceData.Set("status", replicationSet.Status)

	if err := resourceData.Set("region", FlattenRegions(replicationSet.RegionMap)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, resourceData.Id(), err)
	}

	if diagErr := SetResourceDataTags(context, resourceData, meta, client, ResNameReplicationSet); diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceReplicationSetUpdate(context context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	if resourceData.HasChanges("region") {
		input := &ssmincidents.UpdateReplicationSetInput{
			Arn: aws.String(resourceData.Id()),
		}

		if err := updateRegionsInput(resourceData, input); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, resourceData.Id(), err)
		}

		log.Printf("[DEBUG] Updating SSMIncidents ReplicationSet (%s): %#v", resourceData.Id(), input)
		_, err := client.UpdateReplicationSet(context, input)
		if err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, resourceData.Id(), err)
		}

		getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
			Arn: aws.String(resourceData.Id()),
		}

		if err := ssmincidents.NewWaitForReplicationSetActiveWaiter(client).Wait(context, getReplicationSetInput, resourceData.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForUpdate, ResNameReplicationSet, resourceData.Id(), err)
		}

	}

	// tags_all does not detect changes when tag value is "" while this change is detected by tags
	if resourceData.HasChanges("tags_all", "tags") {
		log.Printf("[DEBUG] Updating SSMIncidents ReplicationSet tags")

		if err := UpdateResourceTags(context, client, resourceData); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, resourceData.Id(), err)
		}
	}

	return resourceReplicationSetRead(context, resourceData, meta)
}

func resourceReplicationSetDelete(context context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	log.Printf("[INFO] Deleting SSMIncidents ReplicationSet %s", resourceData.Id())

	_, err := client.DeleteReplicationSet(context, &ssmincidents.DeleteReplicationSetInput{
		Arn: aws.String(resourceData.Id()),
	})

	if err != nil {
		var notFoundError *types.ResourceNotFoundException
		if errors.As(err, &notFoundError) {
			return nil
		}

		return create.DiagError(names.SSMIncidents, create.ErrActionDeleting, ResNameReplicationSet, resourceData.Id(), err)
	}

	getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(resourceData.Id()),
	}

	if err := ssmincidents.NewWaitForReplicationSetDeletedWaiter(client).Wait(context, getReplicationSetInput, resourceData.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForDeletion, ResNameReplicationSet, resourceData.Id(), err)
	}

	return nil
}

//TODO: refactor regionListToMap, updateRegionsInput

// converts a list of regions to a map with the region name as the key and the rest
// of the region data as the values so that it is easier to loop through and process
func regionListToMap(list []interface{}) map[string]map[string]interface{} {
	ret := make(map[string]map[string]interface{})
	for _, val := range list {
		curr := val.(map[string]interface{})
		regionName := curr["name"].(string)
		delete(curr, "name")
		ret[regionName] = curr
	}

	return ret
}

// updates UpdateReplicationSetInput to include any required actions
// invalid updates return errors from AWS Api
func updateRegionsInput(d *schema.ResourceData, in *ssmincidents.UpdateReplicationSetInput) error {
	o, n := d.GetChange("region")
	oldRegions := regionListToMap(o.(*schema.Set).List())
	newRegions := regionListToMap(n.(*schema.Set).List())

	for region, oldVal := range oldRegions {
		if newVal, ok := newRegions[region]; !ok {
			// this region has been destroyed

			action := &types.UpdateReplicationSetActionMemberDeleteRegionAction{
				Value: types.DeleteRegionAction{
					RegionName: aws.String(region),
				},
			}

			in.Actions = append(in.Actions, action)
		} else {
			oldcmk := oldVal["kms_key_arn"].(string)
			newcmk := newVal["kms_key_arn"].(string)

			if oldcmk != newcmk {
				return fmt.Errorf("error: Incident Manager does not support updating Customer Managed Keys. To do this, remove the region, and then re-create it with the new key")
			}
		}
	}

	for region, newVal := range newRegions {
		if _, ok := oldRegions[region]; !ok {
			// this region is newly created
			newcmk := newVal["kms_key_arn"].(string)

			action := &types.UpdateReplicationSetActionMemberAddRegionAction{
				Value: types.AddRegionAction{
					RegionName: aws.String(region),
				},
			}

			if newcmk != "DefaultKey" {
				action.Value.SseKmsKeyId = aws.String(newcmk)
			}

			in.Actions = append(in.Actions, action)
		}
	}

	return nil
}

func resourceReplicationSetImport(context context.Context, resourceData *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	arn, err := GetReplicationSetARN(context, client)

	if err != nil {
		return nil, err
	}

	resourceData.SetId(arn)

	return []*schema.ResourceData{resourceData}, nil
}
