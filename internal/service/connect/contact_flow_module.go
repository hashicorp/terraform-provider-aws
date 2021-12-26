package connect

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/mitchellh/go-homedir"
)

const awsMutexConnectContactFlowModuleKey = `aws_connect_contact_flow_module`

func ResourceContactFlowModule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContactFlowModuleCreate,
		ReadContext:   resourceContactFlowModuleRead,
		UpdateContext: resourceContactFlowModuleUpdate,
		DeleteContext: resourceContactFlowModuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(connectContactFlowModuleCreateTimeout),
			Update: schema.DefaultTimeout(connectContactFlowModuleUpdateTimeout),
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

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, connect.ErrCodeResourceNotFoundException, "") {
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
