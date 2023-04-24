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

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

	out, err := conn.RegisterTargets(ctx, in)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameRegisterTargets, "id", err)
	}

	if out == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameRegisterTargets, "id", errors.New("empty output"))
	}

	d.SetId(aws.ToString(in.TargetGroupIdentifier))
	// d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", d.Get("target_group_identifier"))))

	return resourceRegisterTargetsRead(ctx, d, meta)
}

func resourceRegisterTargetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fmt.Println("I am at read")
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	targetGroupIdentifier := d.Get("target_group_identifier").(string)
	fmt.Println(targetGroupIdentifier)
	out, err := findRegisterTargets(ctx, conn, targetGroupIdentifier)
	fmt.Println(d.Id())
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

	return resourceRegisterTargetsRead(ctx, d, meta)
}

func resourceRegisterTargetsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	log.Printf("[INFO] Deleting VpcLattice RegisterTargets %s", d.Id())
	targetGroupIdentifier := d.Get("target_group_identifier").(string)
	targetsRaw := d.Get("targets").([]interface{})
	targets := expandTargets(targetsRaw)

	_, err := conn.DeregisterTargets(ctx, &vpclattice.DeregisterTargetsInput{
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

	return nil
}

func findRegisterTargets(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.ListTargetsOutput, error) {
	fmt.Println("I am at find")
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

	return out, nil
}

func flattenTargets(apiObjects []types.TargetSummary) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		fmt.Printf("[DEBUG] flattenTargets - apiObject: %#v\n", apiObject)
		tfList = append(tfList, flattenTarget(&apiObject))
	}

	fmt.Printf("[DEBUG] flattenTargets - tfList: %#v\n", tfList)
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
	fmt.Println(tfMap)
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
	j, err := json.Marshal(apiObjects)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Returning from defaultAction:", string(j))
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
