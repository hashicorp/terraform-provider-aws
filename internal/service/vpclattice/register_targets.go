package vpclattice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_vpclattice_register_targets", name="Register Targets")
func ResourceRegisterTargets() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegisterTargetsCreate,
		ReadWithoutTimeout:   resourceRegisterTargetsRead,
		UpdateWithoutTimeout: resourceRegisterTargetsUpdate,
		DeleteWithoutTimeout: resourceRegisterTargetsDelete,

		// Importer: &schema.ResourceImporter{
		// 	StateContext: schema.ImportStatePassthroughContext,
		// },
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("target_group_identifier", d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"target_group_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"targets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

const (
	ResNameRegisterTargets = "Register Targets"
)

func resourceRegisterTargetsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()
	in := &vpclattice.RegisterTargetsInput{
		TargetGroupIdentifier: aws.String(d.Get("target_group_identifier").(string)),
	}

	if v, ok := d.GetOk("targets"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		in.Targets = expandTargets(v.([]interface{}))
	}

	log.Printf("[INFO] Registering Target  with Target Group %s",
		d.Get("target_group_identifier").(string))

	out, err := conn.RegisterTargets(ctx, in)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameRegisterTargets, "id", err)
	}

	if out == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameRegisterTargets, "id", errors.New("empty output"))
	}

	d.SetId(aws.ToString(in.TargetGroupIdentifier))

	if _, err := waitRegisterTargets(ctx, conn, d.Id(), out, d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionWaitingForCreation, ResNameRegisterTargets, d.Id(), err)
	}

	return resourceRegisterTargetsRead(ctx, d, meta)
}

func resourceRegisterTargetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	targetGroupIdentifier := d.Get("target_group_identifier").(string)
	out, err := findRegisterTargets(ctx, conn, targetGroupIdentifier)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VpcLattice RegisterTargets (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, ResNameRegisterTargets, d.Id(), err)
	}

	// if err := d.Set("targets", []interface{}{flattenTargets(out.Items)}); err != nil {
	// 	return create.DiagError(names.VPCLattice, create.ErrActionSetting, ResNameRegisterTargets, d.Id(), err)
	// }
	if err := d.Set("targets", flattenTargets(out.Items)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting targets: %s", err))
	}

	return nil
}

func resourceRegisterTargetsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("targets") {
		conn := meta.(*conns.AWSClient).VPCLatticeClient()

		targetGroupIdentifier := d.Get("target_group_identifier").(string)

		// Deregister old targets
		oldTargetsRaw, _ := d.GetChange("targets")
		oldTargets := expandTargets(oldTargetsRaw.([]interface{}))

		_, err := conn.DeregisterTargets(ctx, &vpclattice.DeregisterTargetsInput{
			TargetGroupIdentifier: aws.String(targetGroupIdentifier),
			Targets:               oldTargets,
		})

		if err != nil {
			return diag.FromErr(fmt.Errorf("error deregistering old targets: %s", err))
		}

		// Register new targets
		newTargetsRaw := d.Get("targets")
		newTargets := expandTargets(newTargetsRaw.([]interface{}))

		_, err = conn.RegisterTargets(ctx, &vpclattice.RegisterTargetsInput{
			TargetGroupIdentifier: aws.String(targetGroupIdentifier),
			Targets:               newTargets,
		})

		if err != nil {
			return diag.FromErr(fmt.Errorf("error registering new targets: %s", err))
		}
	}

	return resourceRegisterTargetsRead(ctx, d, meta)
}

func resourceRegisterTargetsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	log.Printf("[INFO] Deleting VpcLattice RegisterTargets %s", d.Id())
	targetGroupIdentifier := d.Get("target_group_identifier").(string)
	targetsRaw := d.Get("targets").([]interface{})
	targets := expandTargets(targetsRaw)

	out, err := conn.DeregisterTargets(ctx, &vpclattice.DeregisterTargetsInput{
		TargetGroupIdentifier: aws.String(targetGroupIdentifier),
		Targets:               targets,
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.VPCLattice, create.ErrActionDeleting, ResNameRegisterTargets, d.Id(), err)
	}

	if _, err := waitDeleteTargets(ctx, conn, d.Id(), out, d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameTargetGroup, d.Id(), err)
	}

	return nil
}

func findRegisterTargets(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.ListTargetsOutput, error) {
	in := &vpclattice.ListTargetsInput{
		TargetGroupIdentifier: aws.String(id),
	}
	out, err := conn.ListTargets(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Items == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}
	j, err := json.Marshal(out)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Returning from Find Targets:", string(j))
	return out, nil
}

func waitRegisterTargets(ctx context.Context, conn *vpclattice.Client, id string, out *vpclattice.RegisterTargetsOutput, timeout time.Duration) (*vpclattice.RegisterTargetsOutput, error) {
	var lastErr error

	for _, target := range out.Successful {
		stateConf := &retry.StateChangeConf{
			Pending:                   enum.Slice(types.TargetStatusInitial),
			Target:                    enum.Slice(types.TargetStatusHealthy, types.TargetStatusUnhealthy),
			Refresh:                   statusTarget(ctx, conn, id, target),
			Timeout:                   timeout,
			NotFoundChecks:            20,
			ContinuousTargetOccurence: 2,
		}

		outputRaw, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			lastErr = err
			continue
		}

		if out, ok := outputRaw.(*vpclattice.RegisterTargetsOutput); ok {
			out.Successful = append(out.Successful, target)
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return out, nil
}

func waitDeleteTargets(ctx context.Context, conn *vpclattice.Client, id string, out *vpclattice.DeregisterTargetsOutput, timeout time.Duration) (*vpclattice.DeregisterTargetsOutput, error) {
	var lastErr error

	for _, target := range out.Successful {
		stateConf := &retry.StateChangeConf{
			Pending: enum.Slice(types.TargetStatusDraining, types.TargetStatusHealthy, types.TargetStatusUnhealthy),
			Target:  []string{},
			Refresh: statusTarget(ctx, conn, id, target),
			Timeout: timeout,
		}

		outputRaw, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			lastErr = err
			continue
		}

		if out, ok := outputRaw.(*vpclattice.RegisterTargetsOutput); ok {
			out.Successful = append(out.Successful, target)
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return out, nil
}

func statusTarget(ctx context.Context, conn *vpclattice.Client, id string, target types.Target) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		in := &vpclattice.ListTargetsInput{
			TargetGroupIdentifier: aws.String(id),
		}
		out, err := conn.ListTargets(ctx, in)
		if err != nil {
			return nil, "", err
		}

		for _, targetSummary := range out.Items {
			if aws.ToString(targetSummary.Id) == aws.ToString(target.Id) && (target.Port == nil || aws.ToInt32(targetSummary.Port) == aws.ToInt32(target.Port)) {
				return targetSummary, string(targetSummary.Status), nil
			}
		}

		return nil, "", &retry.NotFoundError{LastError: err, LastRequest: in}
	}
}

func flattenTargets(apiObjects []types.TargetSummary) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenTarget(&apiObject))
	}

	return tfList
}

func flattenTarget(apiObject *types.TargetSummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Id; v != nil {
		tfMap["id"] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap["port"] = aws.ToInt32(v)
	}

	return tfMap
}

// Expand function for target_groups
func expandTargets(tfList []interface{}) []types.Target {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Target

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			fmt.Printf("Error: Type assertion failed for tfMapRaw: %v\n", tfMapRaw)
			continue
		}

		apiObject := expandTarget(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandTarget(tfMap map[string]interface{}) types.Target {
	apiObject := types.Target{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
		fmt.Println(apiObject.Id)
	}

	if v, ok := tfMap["port"].(int); ok && v != 0 {
		apiObject.Port = aws.Int32(int32(v))
		fmt.Println(apiObject.Port)
	}

	return apiObject
}
