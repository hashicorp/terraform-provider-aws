package datasync

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAgent() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAgentCreate,
		ReadWithoutTimeout:   resourceAgentRead,
		UpdateWithoutTimeout: resourceAgentUpdate,
		DeleteWithoutTimeout: resourceAgentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"activation_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ExactlyOneOf:  []string{"activation_key", "ip_address"},
				ConflictsWith: []string{"private_link_endpoint"},
			},
			"ip_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"activation_key", "ip_address"},
			},
			"private_link_endpoint": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"activation_key"},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_group_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAgentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	activationKey := d.Get("activation_key").(string)
	agentIpAddress := d.Get("ip_address").(string)

	// Perform one time fetch of activation key from gateway IP address.
	if activationKey == "" {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: time.Second * 10,
		}
		region := meta.(*conns.AWSClient).Region

		var requestURL string
		if v, ok := d.GetOk("private_link_endpoint"); ok {
			requestURL = fmt.Sprintf("http://%s/?gatewayType=SYNC&activationRegion=%s&endpointType=PRIVATE_LINK&privateLinkEndpoint=%s", agentIpAddress, region, v.(string))
		} else {
			requestURL = fmt.Sprintf("http://%s/?gatewayType=SYNC&activationRegion=%s", agentIpAddress, region)
		}

		request, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating HTTP request: %s", err)
		}

		var response *http.Response
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			response, err = client.Do(request)

			if err, ok := err.(net.Error); ok {
				return resource.RetryableError(fmt.Errorf("error making HTTP request: %w", err))
			}

			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("error making HTTP request: %w", err))
			}

			if response == nil {
				return resource.NonRetryableError(fmt.Errorf("no response for activation key request"))
			}

			log.Printf("[DEBUG] Received HTTP response: %#v", response)
			if expected := http.StatusFound; expected != response.StatusCode {
				return resource.NonRetryableError(fmt.Errorf("expected HTTP status code %d, received: %d", expected, response.StatusCode))
			}

			redirectURL, err := response.Location()
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("error extracting HTTP Location header: %w", err))
			}

			if errorType := redirectURL.Query().Get("errorType"); errorType == "PRIVATE_LINK_ENDPOINT_UNREACHABLE" {
				errMessage := fmt.Errorf("got error during activation: %s", errorType)
				return resource.RetryableError(errMessage)
			}

			activationKey = redirectURL.Query().Get("activationKey")

			return nil
		})

		if tfresource.TimedOut(err) {
			return sdkdiag.AppendErrorf(diags, "timeout retrieving activation key from IP Address (%s): %s", agentIpAddress, err)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "retrieving activation key from IP Address (%s): %s", agentIpAddress, err)
		}

		if activationKey == "" {
			return sdkdiag.AppendErrorf(diags, "empty activationKey received from IP Address: %s", agentIpAddress)
		}
	}

	input := &datasync.CreateAgentInput{
		ActivationKey: aws.String(activationKey),
		Tags:          Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("name"); ok {
		input.AgentName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_arns"); ok {
		input.SecurityGroupArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_arns"); ok {
		input.SubnetArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("vpc_endpoint_id"); ok {
		input.VpcEndpointId = aws.String(v.(string))
	}

	output, err := conn.CreateAgentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Agent: %s", err)
	}

	d.SetId(aws.StringValue(output.AgentArn))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return FindAgentByARN(ctx, conn, d.Id())
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DataSync Agent (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceAgentRead(ctx, d, meta)...)
}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindAgentByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Agent (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Agent (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.AgentArn)
	d.Set("name", output.Name)
	if plc := output.PrivateLinkConfig; plc != nil {
		d.Set("private_link_endpoint", plc.PrivateLinkEndpoint)
		d.Set("security_group_arns", flex.FlattenStringList(plc.SecurityGroupArns))
		d.Set("subnet_arns", flex.FlattenStringList(plc.SubnetArns))
		d.Set("vpc_endpoint_id", plc.VpcEndpointId)
	} else {
		d.Set("private_link_endpoint", "")
		d.Set("security_group_arns", nil)
		d.Set("subnet_arns", nil)
		d.Set("vpc_endpoint_id", "")
	}

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DataSync Agent (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceAgentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn()

	if d.HasChange("name") {
		input := &datasync.UpdateAgentInput{
			AgentArn: aws.String(d.Id()),
			Name:     aws.String(d.Get("name").(string)),
		}

		_, err := conn.UpdateAgentWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Agent (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Agent (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAgentRead(ctx, d, meta)...)
}

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn()

	log.Printf("[DEBUG] Deleting DataSync Agent: %s", d.Id())
	_, err := conn.DeleteAgentWithContext(ctx, &datasync.DeleteAgentInput{
		AgentArn: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Agent (%s): %s", d.Id(), err)
	}

	return diags
}
