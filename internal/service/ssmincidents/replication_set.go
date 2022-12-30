package ssmincidents

import (
	// goimports -w <file> fixes these imports.
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

func ResourceReplicationSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationSetCreate,
		ReadWithoutTimeout:   resourceReplicationSetRead,
		UpdateWithoutTimeout: resourceReplicationSetUpdate,
		DeleteWithoutTimeout: resourceReplicationSetDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute), //TODO: recommend customer to set this higher if required
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regions": { // maps each region to an optional SseKmsKeyId, "" represents no customer managed key
				Type:     schema.TypeString,
				Required: true,
				MinItems: 1,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameReplicationSet = "Replication Set"
)

func resourceReplicationSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).SSMIncidentsClient

	in := &ssmincidents.CreateReplicationSetInput{
		Regions:     expandRegions(d.Get("regions").(map[string]string)),
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

	conn := meta.(*conns.AWSClient).SSMIncidentsClient

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

	if err := d.Set("regions", flattenRegions(out.RegionMap)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, d.Id(), err)
	}

	return nil
}

func resourceReplicationSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).SSMIncidentsClient

	if d.HasChanges("regions") {

		diagErr := updateRegions(ctx, conn, d)
		if diagErr != nil {
			return diagErr
		}
	}

	if d.HasChanges("tags_all") {

		o, n := d.GetChange("tags_all")

		if err := updateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}

	}

	return resourceReplicationSetRead(ctx, d, meta)
}

func resourceReplicationSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).SSMIncidentsClient

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

// we finish wait once we receive status is Updating/Deleting since update/Deletion can potentially take very long (5 mins-24 hrs)
// so sometimes there is inconsistency between terraform state and reality
// this behaviour is noted in documentation
func waitReplicationSetUpdated(ctx context.Context, conn *ssmincidents.Client, id string, timeout time.Duration) (*types.ReplicationSet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{string(types.ReplicationSetStatusActive)},
		Target:  []string{string(types.ReplicationSetStatusUpdating)},
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
		Pending: []string{string(types.ReplicationSetStatusActive)},
		Target:  []string{string(types.ReplicationSetStatusDeleting)},
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

func expandRegions(regions map[string]string) map[string]types.RegionMapInputValue {

	ret := make(map[string]types.RegionMapInputValue)
	for k, v := range regions {
		input := types.RegionMapInputValue{}

		if v != "" {
			input.SseKmsKeyId = aws.String(v)
		}

		ret[k] = input
	}

	return ret
}

func flattenRegions(regions map[string]types.RegionInfo) map[string]string {

	ret := make(map[string]string)
	for k, v := range regions {

		if v.SseKmsKeyId == nil {
			ret[k] = ""
		} else {
			ret[k] = aws.ToString(v.SseKmsKeyId)
		}
	}

	return ret
}

// makes api call for updating a single region
func updateRegion(ctx context.Context, conn *ssmincidents.Client, d *schema.ResourceData, action types.UpdateReplicationSetAction) diag.Diagnostics {

	in := &ssmincidents.UpdateReplicationSetInput{
		Arn:         aws.String(d.Id()),
		ClientToken: aws.String(GenerateClientToken()),
	}

	in.Actions = append(in.Actions, action)

	log.Printf("[DEBUG] Updating SSMIncidents ReplicationSet (%s): %#v", d.Id(), in)
	_, err := conn.UpdateReplicationSet(ctx, in)
	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
	}

	if _, err := waitReplicationSetUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForUpdate, ResNameReplicationSet, d.Id(), err)
	}

	return nil

}

// performs all updates to regions field
func updateRegions(ctx context.Context, conn *ssmincidents.Client, d *schema.ResourceData) diag.Diagnostics {

	o, n := d.GetChange("regions")
	oldRegions := o.(map[string]string)
	newRegions := n.(map[string]string)

	if !UpdateRegionsIsValid(newRegions) {
		return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForUpdate, ResNameReplicationSet, d.Id(),
			fmt.Errorf("expected all regions to either have a customer managed key or none of the regions to have a key"))
	}

	// api call only accepts one region update at a time, so we split an update into multiple api calls
	// if a region has changed only its cmk, we must delete and recreate it
	// what if last region? == TODO == figure out this
	//todo: think of other edge cases
	// 1 region -> 3 (now we cannot add all in one go, we just add first one
	// what if we add random combinations with changing keys or some weird shit)
	// changing all from non-cmk to cmk or vice versa, cannot just destroy all of them at once
	// 100% need to make this logic testable

	for region, oldcmk := range oldRegions {
		if newcmk, ok := newRegions[region]; !ok || oldcmk != newcmk {
			// this region has been destroyed
			in := &types.UpdateReplicationSetActionMemberDeleteRegionAction{
				Value: types.DeleteRegionAction{
					RegionName: aws.String(region),
				},
			}

			diagErr := updateRegion(ctx, conn, d, in)

			if diagErr != nil {
				return diagErr
			}
		}
	}

	for region, newcmk := range newRegions {

		if oldcmk, ok := oldRegions[region]; !ok || oldcmk != newcmk {
			// this region is newly created

			in := &types.UpdateReplicationSetActionMemberAddRegionAction{
				Value: types.AddRegionAction{
					RegionName: aws.String(region),
				},
			}

			if newcmk != "" {
				in.Value.SseKmsKeyId = aws.String(newcmk)
			}

			diagErr := updateRegion(ctx, conn, d, in)

			if diagErr != nil {
				return diagErr
			}
		}
	}

	return nil

}

// valid if all region values are "" or all have customer managed keys
// invalid if some regions with empty value and others with proper key
func UpdateRegionsIsValid(newRegions map[string]string) bool {

	s := "INVALIDREGION"

	for _, v := range newRegions {
		if s == "INVALIDREGION" {
			s = v
		} else if s == "" && v != "" || s != "" && v == "" {
			return false
		}
	}

	return true
}

func updateTags(ctx context.Context, conn *ssmincidents.Client, arn string, oldTagsMap interface{}, newTagsMap interface{}) error {

	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &ssmincidents.UntagResourceInput{
			ResourceArn: aws.String(arn),
			TagKeys:     removedTags.Keys(),
		}
		_, err := conn.UntagResource(ctx, input)

		if err != nil {
			return err
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &ssmincidents.TagResourceInput{
			ResourceArn: aws.String(arn),
			Tags:        updatedTags.IgnoreAWS().Map(),
		}

		_, err := conn.TagResource(ctx, input)

		if err != nil {
			return err
		}
	}

	return nil

}

func ListTags(ctx context.Context, conn *ssmincidents.Client, arn string) (tftags.KeyValueTags, error) {
	input := &ssmincidents.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.ListTagsForResource(ctx, input)

	if err != nil {
		return tftags.New(nil), err
	}

	return tftags.New(output.Tags), nil
}
