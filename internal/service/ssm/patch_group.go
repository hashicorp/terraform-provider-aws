package ssm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourcePatchGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePatchGroupCreate,
		ReadWithoutTimeout:   resourcePatchGroupRead,
		DeleteWithoutTimeout: resourcePatchGroupDelete,

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourcePatchGroupV0().CoreConfigSchema().ImpliedType(),
				Upgrade: PatchGroupStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"baseline_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"patch_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePatchGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	baselineId := d.Get("baseline_id").(string)
	patchGroup := d.Get("patch_group").(string)

	params := &ssm.RegisterPatchBaselineForPatchGroupInput{
		BaselineId: aws.String(baselineId),
		PatchGroup: aws.String(patchGroup),
	}

	resp, err := conn.RegisterPatchBaselineForPatchGroupWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering SSM Patch Baseline (%s) for Patch Group (%s): %s", baselineId, patchGroup, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", aws.StringValue(resp.PatchGroup), aws.StringValue(resp.BaselineId)))

	return append(diags, resourcePatchGroupRead(ctx, d, meta)...)
}

func resourcePatchGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	patchGroup, baselineId, err := ParsePatchGroupID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing SSM Patch Group ID (%s): %s", d.Id(), err)
	}

	group, err := FindPatchGroup(ctx, conn, patchGroup, baselineId)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Patch Group (%s): %s", d.Id(), err)
	}

	if group == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading SSM Patch Group (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] SSM Patch Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	var groupBaselineId string
	if group.BaselineIdentity != nil {
		groupBaselineId = aws.StringValue(group.BaselineIdentity.BaselineId)
	}

	d.Set("baseline_id", groupBaselineId)
	d.Set("patch_group", group.PatchGroup)

	return diags
}

func resourcePatchGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	patchGroup, baselineId, err := ParsePatchGroupID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing SSM Patch Group ID (%s): %s", d.Id(), err)
	}

	params := &ssm.DeregisterPatchBaselineForPatchGroupInput{
		BaselineId: aws.String(baselineId),
		PatchGroup: aws.String(patchGroup),
	}

	_, err = conn.DeregisterPatchBaselineForPatchGroupWithContext(ctx, params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deregistering SSM Patch Baseline (%s) for Patch Group (%s): %s", baselineId, patchGroup, err)
	}

	return diags
}

func ParsePatchGroupID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("please make sure ID is in format PATCH_GROUP,BASELINE_ID")
	}

	return parts[0], parts[1], nil
}
