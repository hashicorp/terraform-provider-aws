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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfio "github.com/hashicorp/terraform-provider-aws/internal/io"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const contactFlowModuleMutexKey = `aws_connect_contact_flow_module`

// @SDKResource("aws_connect_contact_flow_module", name="Contact Flow Module")
// @Tags(identifierAttribute="arn")
func resourceContactFlowModule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContactFlowModuleCreate,
		ReadWithoutTimeout:   resourceContactFlowModuleRead,
		UpdateWithoutTimeout: resourceContactFlowModuleUpdate,
		DeleteWithoutTimeout: resourceContactFlowModuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_module_id": {
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 127),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceContactFlowModuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	input := &connect.CreateContactFlowModuleInput{
		Name:       aws.String(name),
		InstanceId: aws.String(instanceID),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filename"); ok {
		v := v.(string)
		// Grab an exclusive lock so that we're only reading one contact flow module into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		conns.GlobalMutexKV.Lock(contactFlowModuleMutexKey)
		defer conns.GlobalMutexKV.Unlock(contactFlowModuleMutexKey)

		contents, err := tfio.ReadFileContents(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Content = aws.String(string(contents))
	} else if v, ok := d.GetOk(names.AttrContent); ok {
		input.Content = aws.String(v.(string))
	}

	output, err := conn.CreateContactFlowModule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Contact Flow Module (%s): %s", name, err)
	}

	id := contactFlowModuleCreateResourceID(instanceID, aws.ToString(output.Id))
	d.SetId(id)

	return append(diags, resourceContactFlowModuleRead(ctx, d, meta)...)
}

func resourceContactFlowModuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, contactFlowModuleID, err := contactFlowModuleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	contactFlowModule, err := findContactFlowModuleByTwoPartKey(ctx, conn, instanceID, contactFlowModuleID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Contact Flow Module (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Contact Flow Module (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, contactFlowModule.Arn)
	d.Set("contact_flow_module_id", contactFlowModule.Id)
	d.Set(names.AttrContent, contactFlowModule.Content)
	d.Set(names.AttrDescription, contactFlowModule.Description)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, contactFlowModule.Name)

	setTagsOut(ctx, contactFlowModule.Tags)

	return diags
}

func resourceContactFlowModuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, contactFlowModuleID, err := contactFlowModuleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges(names.AttrDescription, names.AttrName) {
		input := &connect.UpdateContactFlowModuleMetadataInput{
			ContactFlowModuleId: aws.String(contactFlowModuleID),
			Description:         aws.String(d.Get(names.AttrDescription).(string)),
			InstanceId:          aws.String(instanceID),
			Name:                aws.String(d.Get(names.AttrName).(string)),
		}

		_, err := conn.UpdateContactFlowModuleMetadata(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Contact Flow Module (%s): %s", d.Id(), err)
		}
	}

	if d.HasChanges(names.AttrContent, "content_hash", "filename") {
		input := &connect.UpdateContactFlowModuleContentInput{
			ContactFlowModuleId: aws.String(contactFlowModuleID),
			InstanceId:          aws.String(instanceID),
		}

		if v, ok := d.GetOk("filename"); ok {
			v := v.(string)
			// Grab an exclusive lock so that we're only reading one contact flow module into
			// memory at a time.
			// See https://github.com/hashicorp/terraform/issues/9364
			conns.GlobalMutexKV.Lock(contactFlowModuleMutexKey)
			defer conns.GlobalMutexKV.Unlock(contactFlowModuleMutexKey)

			contents, err := tfio.ReadFileContents(v)
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input.Content = aws.String(string(contents))
		} else if v, ok := d.GetOk(names.AttrContent); ok {
			input.Content = aws.String(v.(string))
		}

		_, updateContentInputErr := conn.UpdateContactFlowModuleContent(ctx, input)

		if updateContentInputErr != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Contact Flow Module content (%s): %s", d.Id(), updateContentInputErr)
		}
	}

	return append(diags, resourceContactFlowModuleRead(ctx, d, meta)...)
}

func resourceContactFlowModuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, contactFlowModuleID, err := contactFlowModuleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Contact Flow Module: %s", d.Id())
	input := connect.DeleteContactFlowModuleInput{
		ContactFlowModuleId: aws.String(contactFlowModuleID),
		InstanceId:          aws.String(instanceID),
	}
	_, err = conn.DeleteContactFlowModule(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Contact Flow Module (%s): %s", d.Id(), err)
	}

	return diags
}

const contactFlowModuleResourceIDSeparator = ":"

func contactFlowModuleCreateResourceID(instanceID, contactFlowID string) string {
	parts := []string{instanceID, contactFlowID}
	id := strings.Join(parts, contactFlowModuleResourceIDSeparator)

	return id
}

func contactFlowModuleParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, contactFlowModuleResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]scontactFlowModuleID", id, contactFlowModuleResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findContactFlowModuleByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, contactFlowModuleID string) (*awstypes.ContactFlowModule, error) {
	input := &connect.DescribeContactFlowModuleInput{
		ContactFlowModuleId: aws.String(contactFlowModuleID),
		InstanceId:          aws.String(instanceID),
	}

	return findContactFlowModule(ctx, conn, input)
}

func findContactFlowModule(ctx context.Context, conn *connect.Client, input *connect.DescribeContactFlowModuleInput) (*awstypes.ContactFlowModule, error) {
	output, err := conn.DescribeContactFlowModule(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ContactFlowModule == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ContactFlowModule, nil
}
