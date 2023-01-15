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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		CreateWithoutTimeout: resourceReplicationSetCreate,
		ReadWithoutTimeout:   resourceReplicationSetRead,
		UpdateWithoutTimeout: resourceReplicationSetUpdate,
		DeleteWithoutTimeout: resourceReplicationSetDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
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

			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				conn := meta.(*conns.AWSClient).SSMIncidentsClient()

				arn, err := getReplicationSetArn(ctx, conn)

				if err != nil {
					return nil, err
				}

				d.SetId(arn)

				if diagErr := GetSetResourceTags(ctx, d, meta, conn, ResNameReplicationSet); diagErr != nil {
					return nil, fmt.Errorf("tags could not be imported")
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReplicationSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMIncidentsClient()

	in := &ssmincidents.CreateReplicationSetInput{
		Regions:     expandRegions(d.Get("region").(*schema.Set).List()),
		ClientToken: aws.String(GenerateClientToken()),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = tags.IgnoreAWS().Map()
	}

	out, err := conn.CreateReplicationSet(ctx, in)
	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionCreating, ResNameReplicationSet, "", err)
	}

	if out == nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionCreating, ResNameReplicationSet, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Arn))

	if _, err := waitReplicationSetCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForCreation, ResNameReplicationSet, d.Id(), err)
	}

	return resourceReplicationSetRead(ctx, d, meta)
}

func resourceReplicationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMIncidentsClient()

	out, err := FindReplicationSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSMIncidents ReplicationSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("created_by", out.CreatedBy)
	d.Set("created_time", out.CreatedTime.String())
	d.Set("deletion_protected", out.DeletionProtected)
	d.Set("last_modified_by", out.LastModifiedBy)
	d.Set("last_modified_time", out.LastModifiedTime.String())
	d.Set("status", out.Status)

	if err := d.Set("region", flattenRegions(out.RegionMap)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, d.Id(), err)
	}

	if diagErr := GetSetResourceTags(ctx, d, meta, conn, ResNameReplicationSet); diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceReplicationSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMIncidentsClient()

	if d.HasChanges("region") {
		in := &ssmincidents.UpdateReplicationSetInput{
			Arn:         aws.String(d.Id()),
			ClientToken: aws.String(GenerateClientToken()),
		}

		if err := updateRegionsInput(d, in); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}

		log.Printf("[DEBUG] Updating SSMIncidents ReplicationSet (%s): %#v", d.Id(), in)
		_, err := conn.UpdateReplicationSet(ctx, in)
		if err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}

		if _, err := waitReplicationSetUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForUpdate, ResNameReplicationSet, d.Id(), err)
		}
	}

	// tags can have a change without tags_all having a change when value of tag is ""
	if d.HasChanges("tags_all", "tags") {
		log.Printf("[DEBUG] Updating SSMIncidents ReplicationSet tags")

		if err := UpdateResourceTags(ctx, conn, d); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}
	}

	return resourceReplicationSetRead(ctx, d, meta)
}

func resourceReplicationSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMIncidentsClient()

	log.Printf("[INFO] Deleting SSMIncidents ReplicationSet %s", d.Id())

	_, err := conn.DeleteReplicationSet(ctx, &ssmincidents.DeleteReplicationSetInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.SSMIncidents, create.ErrActionDeleting, ResNameReplicationSet, d.Id(), err)
	}

	if _, err := waitReplicationSetDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForDeletion, ResNameReplicationSet, d.Id(), err)
	}

	return nil
}

func waitReplicationSetCreated(ctx context.Context, conn *ssmincidents.Client, id string, timeout time.Duration) (*types.ReplicationSet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{string(types.ReplicationSetStatusCreating)},
		Target:  []string{string(types.ReplicationSetStatusActive)},
		Refresh: statusReplicationSet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.ReplicationSet); ok {
		return out, err
	}

	return nil, err
}

func waitReplicationSetUpdated(ctx context.Context, conn *ssmincidents.Client, id string, timeout time.Duration) (*types.ReplicationSet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{string(types.ReplicationSetStatusUpdating)},
		Target:  []string{string(types.ReplicationSetStatusActive)},
		Refresh: statusReplicationSet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.ReplicationSet); ok {
		return out, err
	}

	return nil, err
}

func waitReplicationSetDeleted(ctx context.Context, conn *ssmincidents.Client, id string, timeout time.Duration) (*types.ReplicationSet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{string(types.ReplicationSetStatusDeleting)},
		Target:  []string{},
		Refresh: statusReplicationSet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.ReplicationSet); ok {
		return out, err
	}

	return nil, err
}

func statusReplicationSet(ctx context.Context, conn *ssmincidents.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindReplicationSetByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

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
				return fmt.Errorf("error: modifying the KMS key of a region must be split into two separate updates")
			}
		}
	}

	for region, newVal := range newRegions {
		if _, ok := oldRegions[region]; !ok {
			newcmk := newVal["kms_key_arn"].(string)

			// this region is newly created

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
