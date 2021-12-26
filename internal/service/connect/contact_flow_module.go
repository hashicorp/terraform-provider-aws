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
