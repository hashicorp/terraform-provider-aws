package aws

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func dataSourceAwsStorageGatewayGatewayActivationKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsStorageGatewayGatewayActivationKeyRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"activation_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.SingleIP(),
			},
		},
	}
}

func dataSourceAwsStorageGatewayGatewayActivationKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 10,
	}

	var activationKey string
	activationRegion := meta.(*AWSClient).region
	ipAddress := d.Get("ip_address").(string)

	requestURL := fmt.Sprintf("http://%s/?activationRegion=%s", ipAddress, activationRegion)
	log.Printf("[DEBUG] Creating HTTP request: %s", requestURL)
	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err)
	}

	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		log.Printf("[DEBUG] Making HTTP request: %s", request.URL.String())
		response, err := client.Do(request)
		if err != nil {
			if err, ok := err.(net.Error); ok {
				errMessage := fmt.Errorf("error making HTTP request: %s", err)
				log.Printf("[DEBUG] retryable %s", errMessage)
				return resource.RetryableError(errMessage)
			}
			return resource.NonRetryableError(fmt.Errorf("error making HTTP request: %s", err))
		}

		log.Printf("[DEBUG] Received HTTP response: %#v", response)
		if response.StatusCode != 302 {
			return resource.NonRetryableError(fmt.Errorf("expected HTTP status code 302, received: %d", response.StatusCode))
		}

		redirectURL, err := response.Location()
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error extracting HTTP Location header: %s", err))
		}

		activationKey = redirectURL.Query().Get("activationKey")

		return nil
	})
	if err != nil {
		return fmt.Errorf("error retrieving activation key from IP Address (%s): %s", ipAddress, err)
	}
	if activationKey == "" {
		return fmt.Errorf("empty activationKey received from IP Address: %s", ipAddress)
	}

	d.SetId(activationKey)
	d.Set("activation_key", activationKey)
	d.Set("ip_address", d.Id())

	return nil
}
