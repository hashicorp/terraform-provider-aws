package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/connect/waiter"
	"github.com/mitchellh/go-homedir"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

const awsMutexConnectContactFlowKey = `aws_connect_contact_flow`

func ResourceContactFlow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContactFlowCreate,
		ReadContext:   resourceContactFlowRead,
		UpdateContext: resourceContactFlowUpdate,
		DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.ConnectContactFlowCreateTimeout),
			Update: schema.DefaultTimeout(waiter.ConnectContactFlowUpdateTimeout),
		},
		CustomizeDiff: SetTagsDiff,
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
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
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
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      connect.ContactFlowTypeContactFlow,
				ValidateFunc: validation.StringInSlice(connect.ContactFlowType_Values(), false),
			},
		},
	}
}

func resourceContactFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)

	input := &connect.CreateContactFlowInput{
		Name:       aws.String(name),
		InstanceId: aws.String(instanceID),
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
		conns.GlobalMutexKV.Lock(awsMutexConnectContactFlowKey)
		defer conns.GlobalMutexKV.Unlock(awsMutexConnectContactFlowKey)
		file, err := resourceAwsConnectContactFlowLoadFileContent(filename)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to load %q: %w", filename, err))
		}
		input.Content = aws.String(file)
	} else if v, ok := d.GetOk("content"); ok {
		input.Content = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().ConnectTags()
	}

	output, err := conn.CreateContactFlowWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Contact Flow (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Contact Flow (%s): empty output", name))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.ContactFlowId)))

	return resourceContactFlowRead(ctx, d, meta)
}

func resourceContactFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeContactFlowWithContext(ctx, &connect.DescribeContactFlowInput{
		ContactFlowId: aws.String(contactFlowID),
		InstanceId:    aws.String(instanceID),
	})

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, connect.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Connect Contact Flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.ContactFlow == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow (%s): empty response", d.Id()))
	}

	d.Set("arn", resp.ContactFlow.Arn)
	d.Set("contact_flow_id", resp.ContactFlow.Id)
	d.Set("instance_id", instanceID)
	d.Set("name", resp.ContactFlow.Name)
	d.Set("description", resp.ContactFlow.Description)
	d.Set("type", resp.ContactFlow.Type)
	d.Set("content", resp.ContactFlow.Content)

	tags := keyvaluetags.ConnectKeyValueTags(resp.ContactFlow.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceContactFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(d.Id())

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
			return diag.FromErr(fmt.Errorf("error updating Connect Contact Flow (%s): %w", d.Id(), updateMetadataInputErr))
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
			conns.GlobalMutexKV.Lock(awsMutexConnectContactFlowKey)
			defer conns.GlobalMutexKV.Unlock(awsMutexConnectContactFlowKey)
			file, err := resourceAwsConnectContactFlowLoadFileContent(filename)
			if err != nil {
				return diag.FromErr(fmt.Errorf("unable to load %q: %w", filename, err))
			}
			updateContentInput.Content = aws.String(file)
		} else if v, ok := d.GetOk("content"); ok {
			updateContentInput.Content = aws.String(v.(string))
		}

		_, updateContentInputErr := conn.UpdateContactFlowContentWithContext(ctx, updateContentInput)

		if updateContentInputErr != nil {
			return diag.FromErr(fmt.Errorf("error updating Connect Contact Flow content (%s): %w", d.Id(), updateContentInputErr))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.ConnectUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceContactFlowRead(ctx, d, meta)
}

//Contact Flows do not support deletion today. We will NoOp the Delete method. Users can rename their flows manually if they want.
// func resourceContactFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	conn := meta.(*conns.AWSClient).ConnectConn

// 	instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(d.Id())

// 	if err != nil {
// 		return diag.FromErr(err)
// 	}

// 	input := &connect.UpdateContactFlowNameInput{
// 		ContactFlowId: aws.String(contactFlowID),
// 		InstanceId:    aws.String(instanceID),
// 		Name:          aws.String(fmt.Sprintf("%s:%s:%d", "zzTrash", d.Get("name").(string), time.Now().Unix())),
// 		Description:   aws.String("DELETED"),
// 	}

// 	_, delerr := conn.UpdateContactFlowNameWithContext(ctx, input)

// 	if delerr != nil {
// 		return diag.FromErr(fmt.Errorf("Unable to delete contact flow: %s", delerr))
// 	}

// 	return nil
// }

func resourceAwsConnectContactFlowParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:contactFlowID", id)
	}

	return parts[0], parts[1], nil
}

func resourceAwsConnectContactFlowLoadFileContent(filename string) (string, error) {
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
