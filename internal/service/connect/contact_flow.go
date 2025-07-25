// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfio "github.com/hashicorp/terraform-provider-aws/internal/io"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const contactFlowMutexKey = `aws_connect_contact_flow`

// @SDKResource("aws_connect_contact_flow", name="Contact Flow")
// @Tags(identifierAttribute="arn")
func resourceContactFlow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContactFlowCreate,
		ReadWithoutTimeout:   resourceContactFlowRead,
		UpdateWithoutTimeout: resourceContactFlowUpdate,
		DeleteWithoutTimeout: resourceContactFlowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				ConflictsWith:    []string{"filename"},
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"content_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filename": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrContent},
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.ContactFlowTypeContactFlow,
				ValidateDiagFunc: enum.Validate[awstypes.ContactFlowType](),
			},
		},
	}
}

func resourceContactFlowCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	input := &connect.CreateContactFlowInput{
		Name:       aws.String(name),
		InstanceId: aws.String(instanceID),
		Tags:       getTagsIn(ctx),
		Type:       awstypes.ContactFlowType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filename"); ok {
		v := v.(string)
		// Grab an exclusive lock so that we're only reading one contact flow into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		conns.GlobalMutexKV.Lock(contactFlowMutexKey)
		defer conns.GlobalMutexKV.Unlock(contactFlowMutexKey)

		contents, err := tfio.ReadFileContents(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Content = aws.String(string(contents))
	} else if v, ok := d.GetOk(names.AttrContent); ok {
		input.Content = aws.String(v.(string))
	}

	output, err := conn.CreateContactFlow(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Contact Flow (%s): %s", name, err)
	}

	id := contactFlowCreateResourceID(instanceID, aws.ToString(output.ContactFlowId))
	d.SetId(id)

	return append(diags, resourceContactFlowRead(ctx, d, meta)...)
}

func resourceContactFlowRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, contactFlowID, err := contactFlowParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	contactFlow, err := findContactFlowByTwoPartKey(ctx, conn, instanceID, contactFlowID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Contact Flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Contact Flow (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, contactFlow.Arn)
	d.Set("contact_flow_id", contactFlow.Id)
	d.Set(names.AttrContent, contactFlow.Content)
	d.Set(names.AttrDescription, contactFlow.Description)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, contactFlow.Name)
	d.Set(names.AttrType, contactFlow.Type)

	setTagsOut(ctx, contactFlow.Tags)

	return diags
}

func resourceContactFlowUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, contactFlowID, err := contactFlowParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges(names.AttrDescription, names.AttrName) {
		input := &connect.UpdateContactFlowNameInput{
			ContactFlowId: aws.String(contactFlowID),
			Description:   aws.String(d.Get(names.AttrDescription).(string)),
			InstanceId:    aws.String(instanceID),
			Name:          aws.String(d.Get(names.AttrName).(string)),
		}

		_, err := conn.UpdateContactFlowName(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Contact Flow (%s): %s", d.Id(), err)
		}
	}

	if d.HasChanges(names.AttrContent, "content_hash", "filename") {
		input := &connect.UpdateContactFlowContentInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
		}

		if v, ok := d.GetOk("filename"); ok {
			v := v.(string)
			// Grab an exclusive lock so that we're only reading one contact flow into
			// memory at a time.
			// See https://github.com/hashicorp/terraform/issues/9364
			conns.GlobalMutexKV.Lock(contactFlowMutexKey)
			defer conns.GlobalMutexKV.Unlock(contactFlowMutexKey)

			contents, err := tfio.ReadFileContents(v)
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input.Content = aws.String(string(contents))
		} else if v, ok := d.GetOk(names.AttrContent); ok {
			input.Content = aws.String(v.(string))
		}

		_, err := conn.UpdateContactFlowContent(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Contact Flow content (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceContactFlowRead(ctx, d, meta)...)
}

func resourceContactFlowDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, contactFlowID, err := contactFlowParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Contact Flow: %s", d.Id())
	input := connect.DeleteContactFlowInput{
		ContactFlowId: aws.String(contactFlowID),
		InstanceId:    aws.String(instanceID),
	}
	_, err = conn.DeleteContactFlow(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Contact Flow (%s): %s", d.Id(), err)
	}

	return diags
}

const contactFlowResourceIDSeparator = ":"

func contactFlowCreateResourceID(instanceID, contactFlowID string) string {
	parts := []string{instanceID, contactFlowID}
	id := strings.Join(parts, contactFlowResourceIDSeparator)

	return id
}

func contactFlowParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, contactFlowResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]scontactFlowID", id, contactFlowResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findContactFlowByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, contactFlowID string) (*awstypes.ContactFlow, error) {
	input := &connect.DescribeContactFlowInput{
		ContactFlowId: aws.String(contactFlowID),
		InstanceId:    aws.String(instanceID),
	}

	return findContactFlow(ctx, conn, input)
}

func findContactFlow(ctx context.Context, conn *connect.Client, input *connect.DescribeContactFlowInput) (*awstypes.ContactFlow, error) {
	output, err := conn.DescribeContactFlow(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ContactFlow == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ContactFlow, nil
}
