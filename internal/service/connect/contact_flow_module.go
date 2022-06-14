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
	"github.com/mitchellh/go-homedir"
)

const contactFlowModuleMutexKey = `aws_connect_contact_flow_module`

func ResourceContactFlowModule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContactFlowModuleCreate,
		ReadContext:   resourceContactFlowModuleRead,
		UpdateContext: resourceContactFlowModuleUpdate,
		DeleteContext: resourceContactFlowModuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_module_id": {
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 127),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceContactFlowModuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)

	input := &connect.CreateContactFlowModuleInput{
		Name:       aws.String(name),
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filename"); ok {
		filename := v.(string)
		// Grab an exclusive lock so that we're only reading one contact flow module into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		conns.GlobalMutexKV.Lock(contactFlowModuleMutexKey)
		defer conns.GlobalMutexKV.Unlock(contactFlowModuleMutexKey)
		file, err := resourceContactFlowModuleLoadFileContent(filename)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to load %q: %w", filename, err))
		}
		input.Content = aws.String(file)
	} else if v, ok := d.GetOk("content"); ok {
		input.Content = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateContactFlowModuleWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Contact Flow Module (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Contact Flow Module (%s): empty output", name))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.Id)))

	return resourceContactFlowModuleRead(ctx, d, meta)
}

func resourceContactFlowModuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID, contactFlowModuleID, err := ContactFlowModuleParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeContactFlowModuleWithContext(ctx, &connect.DescribeContactFlowModuleInput{
		ContactFlowModuleId: aws.String(contactFlowModuleID),
		InstanceId:          aws.String(instanceID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Contact Flow Module (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow Module (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.ContactFlowModule == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow Module (%s): empty response", d.Id()))
	}

	d.Set("arn", resp.ContactFlowModule.Arn)
	d.Set("contact_flow_module_id", resp.ContactFlowModule.Id)
	d.Set("instance_id", instanceID)
	d.Set("name", resp.ContactFlowModule.Name)
	d.Set("description", resp.ContactFlowModule.Description)
	d.Set("content", resp.ContactFlowModule.Content)

	tags := KeyValueTags(resp.ContactFlowModule.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceContactFlowModuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, contactFlowModuleID, err := ContactFlowModuleParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("name", "description") {
		updateMetadataInput := &connect.UpdateContactFlowModuleMetadataInput{
			ContactFlowModuleId: aws.String(contactFlowModuleID),
			Description:         aws.String(d.Get("description").(string)),
			InstanceId:          aws.String(instanceID),
			Name:                aws.String(d.Get("name").(string)),
		}

		_, updateMetadataInputErr := conn.UpdateContactFlowModuleMetadataWithContext(ctx, updateMetadataInput)

		if updateMetadataInputErr != nil {
			return diag.FromErr(fmt.Errorf("error updating Connect Contact Flow Module (%s): %w", d.Id(), updateMetadataInputErr))
		}
	}

	if d.HasChanges("content", "content_hash", "filename") {
		updateContentInput := &connect.UpdateContactFlowModuleContentInput{
			ContactFlowModuleId: aws.String(contactFlowModuleID),
			InstanceId:          aws.String(instanceID),
		}

		if v, ok := d.GetOk("filename"); ok {
			filename := v.(string)
			// Grab an exclusive lock so that we're only reading one contact flow module into
			// memory at a time.
			// See https://github.com/hashicorp/terraform/issues/9364
			conns.GlobalMutexKV.Lock(contactFlowModuleMutexKey)
			defer conns.GlobalMutexKV.Unlock(contactFlowModuleMutexKey)
			file, err := resourceContactFlowModuleLoadFileContent(filename)
			if err != nil {
				return diag.FromErr(fmt.Errorf("unable to load %q: %w", filename, err))
			}
			updateContentInput.Content = aws.String(file)
		} else if v, ok := d.GetOk("content"); ok {
			updateContentInput.Content = aws.String(v.(string))
		}

		_, updateContentInputErr := conn.UpdateContactFlowModuleContentWithContext(ctx, updateContentInput)

		if updateContentInputErr != nil {
			return diag.FromErr(fmt.Errorf("error updating Connect Contact Flow Module content (%s): %w", d.Id(), updateContentInputErr))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceContactFlowModuleRead(ctx, d, meta)
}

func resourceContactFlowModuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, contactFlowModuleID, err := ContactFlowModuleParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Deleting Connect Contact Flow Module : %s", contactFlowModuleID)
	input := &connect.DeleteContactFlowModuleInput{
		ContactFlowModuleId: aws.String(contactFlowModuleID),
		InstanceId:          aws.String(instanceID),
	}

	_, deleteContactFlowModuleErr := conn.DeleteContactFlowModuleWithContext(ctx, input)
	if deleteContactFlowModuleErr != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Contact Flow Module (%s): %w", d.Id(), deleteContactFlowModuleErr))
	}
	return nil
}

func ContactFlowModuleParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:contactFlowModuleID", id)
	}

	return parts[0], parts[1], nil
}

func resourceContactFlowModuleLoadFileContent(filename string) (string, error) {
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
