// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/mitchellh/go-homedir"
)

const contactFlowMutexKey = `aws_connect_contact_flow`

// @SDKResource("aws_connect_contact_flow", name="Contact Flow")
// @Tags(identifierAttribute="arn")
func ResourceContactFlow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContactFlowCreate,
		ReadWithoutTimeout:   resourceContactFlowRead,
		UpdateWithoutTimeout: resourceContactFlowUpdate,
		DeleteWithoutTimeout: resourceContactFlowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				ConflictsWith:    []string{"filename"},
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"content_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filename": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"content"},
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      connect.ContactFlowTypeContactFlow,
				ValidateFunc: validation.StringInSlice(connect.ContactFlowType_Values(), false),
			},
		},
	}
}

func resourceContactFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)

	input := &connect.CreateContactFlowInput{
		Name:       aws.String(name),
		InstanceId: aws.String(instanceID),
		Tags:       getTagsIn(ctx),
		Type:       aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filename"); ok {
		filename := v.(string)
		// Grab an exclusive lock so that we're only reading one contact flow into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		conns.GlobalMutexKV.Lock(contactFlowMutexKey)
		defer conns.GlobalMutexKV.Unlock(contactFlowMutexKey)
		file, err := resourceContactFlowLoadFileContent(filename)
		if err != nil {
			return diag.Errorf("unable to load %q: %s", filename, err)
		}
		input.Content = aws.String(file)
	} else if v, ok := d.GetOk("content"); ok {
		input.Content = aws.String(v.(string))
	}

	output, err := conn.CreateContactFlowWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Connect Contact Flow (%s): %s", name, err)
	}

	if output == nil {
		return diag.Errorf("creating Connect Contact Flow (%s): empty output", name)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.ContactFlowId)))

	return resourceContactFlowRead(ctx, d, meta)
}

func resourceContactFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, contactFlowID, err := ContactFlowParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeContactFlowWithContext(ctx, &connect.DescribeContactFlowInput{
		ContactFlowId: aws.String(contactFlowID),
		InstanceId:    aws.String(instanceID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Contact Flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Connect Contact Flow (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.ContactFlow == nil {
		return diag.Errorf("getting Connect Contact Flow (%s): empty response", d.Id())
	}

	d.Set("arn", resp.ContactFlow.Arn)
	d.Set("contact_flow_id", resp.ContactFlow.Id)
	d.Set("instance_id", instanceID)
	d.Set("name", resp.ContactFlow.Name)
	d.Set("description", resp.ContactFlow.Description)
	d.Set("type", resp.ContactFlow.Type)
	d.Set("content", resp.ContactFlow.Content)

	setTagsOut(ctx, resp.ContactFlow.Tags)

	return nil
}

func resourceContactFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, contactFlowID, err := ContactFlowParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("name", "description") {
		updateMetadataInput := &connect.UpdateContactFlowNameInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
			Name:          aws.String(d.Get("name").(string)),
			Description:   aws.String(d.Get("description").(string)),
		}

		_, updateMetadataInputErr := conn.UpdateContactFlowNameWithContext(ctx, updateMetadataInput)

		if updateMetadataInputErr != nil {
			return diag.Errorf("updating Connect Contact Flow (%s): %s", d.Id(), updateMetadataInputErr)
		}
	}

	if d.HasChanges("content", "content_hash", "filename") {
		updateContentInput := &connect.UpdateContactFlowContentInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
		}

		if v, ok := d.GetOk("filename"); ok {
			filename := v.(string)
			// Grab an exclusive lock so that we're only reading one contact flow into
			// memory at a time.
			// See https://github.com/hashicorp/terraform/issues/9364
			conns.GlobalMutexKV.Lock(contactFlowMutexKey)
			defer conns.GlobalMutexKV.Unlock(contactFlowMutexKey)
			file, err := resourceContactFlowLoadFileContent(filename)
			if err != nil {
				return diag.Errorf("unable to load %q: %s", filename, err)
			}
			updateContentInput.Content = aws.String(file)
		} else if v, ok := d.GetOk("content"); ok {
			updateContentInput.Content = aws.String(v.(string))
		}

		_, updateContentInputErr := conn.UpdateContactFlowContentWithContext(ctx, updateContentInput)

		if updateContentInputErr != nil {
			return diag.Errorf("updating Connect Contact Flow content (%s): %s", d.Id(), updateContentInputErr)
		}
	}

	return resourceContactFlowRead(ctx, d, meta)
}

func resourceContactFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, contactFlowID, err := ContactFlowParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting Connect Contact Flow : %s", contactFlowID)

	input := &connect.DeleteContactFlowInput{
		ContactFlowId: aws.String(contactFlowID),
		InstanceId:    aws.String(instanceID),
	}

	_, deleteContactFlowErr := conn.DeleteContactFlowWithContext(ctx, input)

	if deleteContactFlowErr != nil {
		return diag.Errorf("deleting Connect Contact Flow (%s): %s", d.Id(), deleteContactFlowErr)
	}

	return nil
}

func ContactFlowParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:contactFlowID", id)
	}

	return parts[0], parts[1], nil
}

func resourceContactFlowLoadFileContent(filename string) (string, error) {
	filename, err := homedir.Expand(filename)
	if err != nil {
		return "", err
	}
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}
