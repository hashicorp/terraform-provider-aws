package aws

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/datasync/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/datasync/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAgent() *schema.Resource {
	return &schema.Resource{
		Create: resourceAgentCreate,
		Read:   resourceAgentRead,
		Update: resourceAgentUpdate,
		Delete: resourceAgentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceAgentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	activationKey := d.Get("activation_key").(string)
	agentIpAddress := d.Get("ip_address").(string)

	// Perform one time fetch of activation key from gateway IP address
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
			return fmt.Errorf("error creating HTTP request: %w", err)
		}

		var response *http.Response
		err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			log.Printf("[DEBUG] Making HTTP request: %s", request.URL.String())
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
			return fmt.Errorf("timeout retrieving activation key from IP Address (%s): %w", agentIpAddress, err)
		}

		if err != nil {
			return fmt.Errorf("error retrieving activation key from IP Address (%s): %w", agentIpAddress, err)
		}

		if activationKey == "" {
			return fmt.Errorf("empty activationKey received from IP Address: %s", agentIpAddress)
		}
	}

	input := &datasync.CreateAgentInput{
		ActivationKey: aws.String(activationKey),
		Tags:          tags.IgnoreAws().DatasyncTags(),
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

	log.Printf("[DEBUG] Creating DataSync Agent: %s", input)
	output, err := conn.CreateAgent(input)

	if err != nil {
		return fmt.Errorf("error creating DataSync Agent: %w", err)
	}

	d.SetId(aws.StringValue(output.AgentArn))

	// Agent activations can take a few minutes
	if _, err := waiter.AgentReady(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for DataSync Agent (%s) creation: %s", d.Id(), err)
	}

	return resourceAgentRead(d, meta)
}

func resourceAgentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := finder.AgentByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Agent (%s)not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Agent (%s): %w", d.Id(), err)
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

	tags, err := tftags.DatasyncListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Agent (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAgentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	if d.HasChange("name") {
		input := &datasync.UpdateAgentInput{
			AgentArn: aws.String(d.Id()),
			Name:     aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating DataSync Agent: %s", input)
		_, err := conn.UpdateAgent(input)

		if err != nil {
			return fmt.Errorf("error updating DataSync Agent (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.DatasyncUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Agent (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAgentRead(d, meta)
}

func resourceAgentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	log.Printf("[DEBUG] Deleting DataSync Agent: %s", d.Id())
	_, err := conn.DeleteAgent(&datasync.DeleteAgentInput{
		AgentArn: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "does not exist") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Agent (%s): %w", d.Id(), err)
	}

	return nil
}
