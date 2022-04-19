package lightsail

import (
	"context"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceContainerService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContainerServiceCreate,
		ReadContext:   resourceContainerServiceRead,
		UpdateContext: resourceContainerServiceUpdate,
		DeleteContext: resourceContainerServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotWhiteSpace,
									},
									"image": {
										Type:     schema.TypeString,
										Required: true,
									},
									"command": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"environment": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"port": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port_number": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"protocol": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(lightsail.ContainerServiceProtocol_Values(), false),
												},
											},
										},
									},
								},
							},
						},
						"public_endpoint": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"container_port": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"health_check": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"healthy_threshold": {
													Type:     schema.TypeInt,
													Optional: true,
													Default:  2,
												},
												"interval_seconds": {
													Type:         schema.TypeInt,
													Optional:     true,
													Default:      5,
													ValidateFunc: validation.IntBetween(5, 300),
												},
												"path": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "/",
												},
												"success_codes": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "200-499",
												},
												"timeout_seconds": {
													Type:         schema.TypeInt,
													Optional:     true,
													Default:      2,
													ValidateFunc: validation.IntBetween(2, 60),
												},
												"unhealthy_threshold": {
													Type:     schema.TypeInt,
													Optional: true,
													Default:  2,
												},
											},
										},
									},
								},
							},
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"is_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,2}|[a-z0-9][a-z0-9-]+[a-z0-9]$`), ""),
				),
			},
			"power": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(lightsail.ContainerServicePowerName_Values(), false),
			},
			"power_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"principal_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_domain_names": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"domain_names": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			"resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scale": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 20),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceContainerServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	serviceName := aws.String(d.Get("name").(string))

	containerServiceInput := lightsail.CreateContainerServiceInput{
		ServiceName: serviceName,
		Power:       aws.String(d.Get("power").(string)),
		Scale:       aws.Int64(int64(d.Get("scale").(int))),
	}

	if v, ok := d.GetOk("public_domain_names"); ok {
		containerServiceInput.PublicDomainNames = expandLightsailContainerServicePublicDomainNames(v.([]interface{}))
	}

	if len(tags) > 0 {
		containerServiceInput.Tags = Tags(tags.IgnoreAWS())
	}

	if _, err := conn.CreateContainerService(&containerServiceInput); err != nil {
		log.Printf("[ERROR] Lightsail Container Service (%s) create failed: %s", aws.StringValue(serviceName), err)
		return diag.FromErr(err)
	}

	d.SetId(d.Get("name").(string))
	log.Printf("[INFO] Lightsail Container Service (%s) CreateContainerService call successful, now waiting for ContainerServiceState change", d.Id())

	// first wait for the completion of creating an empty container service
	stateChangeConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.ContainerServiceStatePending},
		Target:     []string{lightsail.ContainerServiceStateReady},
		Refresh:    lightsailContainerServiceRefreshFunc(containerServiceInput.ServiceName, meta),
		Timeout:    25 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil {
		log.Printf("[ERROR] Lightsail Container Service (%s) error waiting for container service to be created: %s", d.Id(), err)
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Lightsail Container Service (%s) create successful", d.Id())

	// Now with a READY and empty container service, do deployment if there is one
	if v, ok := d.GetOk("deployment"); ok {
		deployment := expandLightsailContainerServiceDeployment(v.([]interface{}))

		containerServiceDeploymentInput := lightsail.CreateContainerServiceDeploymentInput{
			ServiceName:    serviceName,
			Containers:     deployment.Containers,
			PublicEndpoint: deployment.PublicEndpoint,
		}

		if _, err := conn.CreateContainerServiceDeployment(&containerServiceDeploymentInput); err != nil {
			log.Printf("[ERROR] Lightsail Container Service (%s) create successful, but deployment failed.", d.Id())
			return diag.FromErr(err)
		}

		// Wait for deployment to success/fail (corresponding to RUNNING/READY)
		// The delay is important to set to be long enough to first let the container service
		// to move out of READY state after a deployment request is sent,
		// because READY is used in Target to detect a failed deployment
		stateChangeConf := &resource.StateChangeConf{
			Pending:    []string{lightsail.ContainerServiceStateDeploying},
			Target:     []string{lightsail.ContainerServiceStateRunning, lightsail.ContainerServiceStateReady},
			Refresh:    lightsailContainerServiceRefreshFunc(containerServiceInput.ServiceName, meta),
			Timeout:    25 * time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		result, err := stateChangeConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[ERROR] Lightsail Container Service (%s) error waiting for container service to deploy: %s", d.Id(), err)
			return diag.FromErr(err)
		}

		r := result.(*lightsail.ContainerService)
		switch aws.StringValue(r.State) {
		case lightsail.ContainerServiceStateRunning:
			log.Printf("[INFO] Lightsail Container Service (%s) deployment successful", d.Id())

		case lightsail.ContainerServiceStateReady:
			log.Printf("[ERROR] Lightsail Container Service (%s) deployment failed", d.Id())
			return diag.Errorf("Lightsail Container Service deployment failed")

		default:
			log.Printf("[ERROR] Lightsail Container Service (%s) in an unexpected state (%s) after deployment", d.Id(), aws.StringValue(r.State))
			return diag.Errorf("Lightsail Container Service deployment failed")
		}
	}

	// once container service creation and/or deployment successful (now enabled by default), disable it if "is_disabled" is true
	if isDisabled := d.Get("is_disabled"); isDisabled.(bool) {
		updateContainerServiceInput := lightsail.UpdateContainerServiceInput{
			ServiceName: serviceName,
			IsDisabled:  aws.Bool(true),
		}

		if _, err := conn.UpdateContainerService(&updateContainerServiceInput); err != nil {
			log.Printf("[ERROR] Lightsail Container Service (%s) create and/or deployment successful, but disabling it failed: %s", d.Id(), err)
			return diag.FromErr(err)
		}

		stateChangeConf := &resource.StateChangeConf{
			Pending:    []string{lightsail.ContainerServiceStateUpdating},
			Target:     []string{lightsail.ContainerServiceStateDisabled},
			Refresh:    lightsailContainerServiceRefreshFunc(containerServiceInput.ServiceName, meta),
			Timeout:    25 * time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil {
			log.Printf("[ERROR] Lightsail Container Service (%s) error waiting for container service to disable: %s", d.Id(), err)
			return diag.FromErr(err)
		}
	}

	return resourceContainerServiceRead(ctx, d, meta)
}

func resourceContainerServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetContainerServices(
		&lightsail.GetContainerServicesInput{
			// d.Id() instead of d.Get("name") used here, because we need to tell Importer how to import
			ServiceName: aws.String(d.Id()),
		},
	)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == lightsail.ErrCodeNotFoundException {
				log.Printf("[WARN] Lightsail Container Service (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		}
		log.Printf("[ERROR] Lightsail Container Service (%s) read failed: %s", d.Id(), err)
		return diag.FromErr(err)
	}

	// just look at index 0 because we only looked up 1 container service
	cs := resp.ContainerServices[0]

	d.Set("name", cs.ContainerServiceName)
	d.Set("power", cs.Power)
	d.Set("scale", cs.Scale)
	d.Set("is_disabled", cs.IsDisabled)
	if err := d.Set("deployment", flattenLightsailContainerServiceDeployment(cs.CurrentDeployment)); err != nil {
		return diag.Errorf("error setting deployment for Lightsail Container Service (%s): %s", d.Id(), err)
	}
	if err := d.Set("public_domain_names", flattenLightsailContainerServicePublicDomainNames(cs.PublicDomainNames)); err != nil {
		return diag.Errorf("error setting public_domain_names for Lightsail Container Service (%s): %s", d.Id(), err)
	}
	d.Set("arn", cs.Arn)
	d.Set("availability_zone", cs.Location.AvailabilityZone)
	d.Set("created_at", aws.TimeValue(cs.CreatedAt).Format(time.RFC3339))
	d.Set("power_id", cs.PowerId)
	d.Set("principal_arn", cs.PrincipalArn)
	d.Set("private_domain_name", cs.PrivateDomainName)
	d.Set("resource_type", cs.ResourceType)
	d.Set("state", cs.State)
	d.Set("url", cs.Url)

	tags := KeyValueTags(cs.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}
	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceContainerServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Id())
	o, n := d.GetChange("is_disabled")
	oldIsDisabled := o.(bool)
	newIsDisabled := n.(bool)

	// fields handled by LightsailUpdateTags, can update no matter container service disabled or not
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			log.Printf("[ERROR] Lightsail Container Service (%s) updating tags failed: %s", d.Id(), err)
			return diag.FromErr(err)
		}

		log.Printf("[INFO] Lightsail Container Service (%s) tags update successful", d.Id())
	}

	// fields handled by CreateContainerServiceDeployment
	if deployment, deploymentChanged := lightsailContainerServiceDeploymentChanged(d); deploymentChanged {
		// if the change is the removal of deployment, it is not allowed by Lightsail
		// a container service cannot go back to READY (i.e. empty container service) once a previous deployment was successful
		if deployment == nil {
			return diag.Errorf("a container service's deployment cannot be removed once a previous deployment was successful. " +
				"You must specify a deployment now.")
		}

		// if the container service is disabled, it must be enabled before handling changes on deployment
		publicDomainNames, publicDomainNamesChanged := lightsailContainerServicePublicDomainNamesChanged(d)
		if oldIsDisabled || d.HasChanges("power", "scale") || publicDomainNamesChanged {
			updateContainerServiceInput := lightsail.UpdateContainerServiceInput{
				ServiceName:       serviceName,
				IsDisabled:        aws.Bool(false),
				Power:             aws.String(d.Get("power").(string)),
				PublicDomainNames: publicDomainNames,
				Scale:             aws.Int64(int64(d.Get("scale").(int))),
			}

			if _, err := conn.UpdateContainerService(&updateContainerServiceInput); err != nil {
				log.Printf("[ERROR] Lightsail Container Service (%s) update failed: %s", d.Id(), err)
				return diag.FromErr(err)
			}

			stateChangeConf := &resource.StateChangeConf{
				Pending:    []string{lightsail.ContainerServiceStateUpdating},
				Target:     []string{lightsail.ContainerServiceStateReady, lightsail.ContainerServiceStateRunning},
				Refresh:    lightsailContainerServiceRefreshFunc(serviceName, meta),
				Timeout:    25 * time.Minute,
				Delay:      5 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil {
				log.Printf("[ERROR] Lightsail Container Service (%s) error waiting for container service to update: %s", d.Id(), err)
				return diag.FromErr(err)
			}
		}

		// now that with an enabled and updated container service, we can handle changes on deployment
		createContainerServiceDeploymentInput := lightsail.CreateContainerServiceDeploymentInput{
			Containers:     deployment.Containers,
			PublicEndpoint: deployment.PublicEndpoint,
			ServiceName:    serviceName,
		}

		if _, err := conn.CreateContainerServiceDeployment(&createContainerServiceDeploymentInput); err != nil {
			log.Printf("[ERROR] Lightsail Container Service (%s) update deployment failed: %s", d.Id(), err)
			return diag.FromErr(err)
		}

		// the delay is important to set to be long enough to first let the container service
		// to move out of READY/RUNNING state after a deployment request is sent,
		// because READY/RUNNING is used in Target
		stateChangeConf := &resource.StateChangeConf{
			Pending:    []string{lightsail.ContainerServiceStateDeploying},
			Target:     []string{lightsail.ContainerServiceStateReady, lightsail.ContainerServiceStateRunning},
			Refresh:    lightsailContainerServiceRefreshFunc(serviceName, meta),
			Timeout:    25 * time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		resp, err := stateChangeConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[ERROR] Lightsail Container Service (%s) error waiting for container service to deploy: %s", d.Id(), err)
			return diag.FromErr(err)
		}

		// see if the deployment succeeds or not
		container := resp.(*lightsail.ContainerService)
		switch aws.StringValue(container.State) {
		// READY means the new deployment failed, because the container service is back to being as if an empty container service
		case lightsail.ContainerServiceStateReady:
			log.Printf("[ERROR] Lightsail Container Service (%s) deployment failed", d.Id())
			return diag.Errorf("Lightsail Container Service deployment failed")

		// RUNNING does not necessarily mean the newest deployment is successful
		case lightsail.ContainerServiceStateRunning:
			req := lightsail.GetContainerServiceDeploymentsInput{
				ServiceName: serviceName,
			}
			deploymentsOutput, err := conn.GetContainerServiceDeployments(&req)
			if err != nil {
				log.Printf("[ERROR] Lightsail Container Service (%s) get deployment status failed: %s", d.Id(), err)
				return diag.FromErr(err)
			}

			// look at the first deployment of the response, which is the new deployment
			latestDeployment := deploymentsOutput.Deployments[0]
			switch aws.StringValue(latestDeployment.State) {
			// FAILED means the new deployment failed, and the RUNNING deployment is a previous successful deployment
			case lightsail.ContainerServiceDeploymentStateFailed:
				log.Printf("[ERROR] Lightsail Container Service (%s) deployment failed", d.Id())
				return diag.Errorf("Lightsail Container Service deployment failed")

			// ACTIVE means the new deployment is successful
			case lightsail.ContainerServiceDeploymentStateActive:
				log.Printf("[INFO] Lightsail Container Service (%s) deployment successful", d.Id())

			default:
				log.Printf("[ERROR] Lightsail Container Service (%s) deployment in an unexpected state (%s)", d.Id(), aws.StringValue(latestDeployment.State))
				return diag.Errorf("Lightsail Container Service deployment failed")
			}

		default:
			log.Printf("[ERROR] Lightsail Container Service (%s) in an unexpected state (%s)", d.Id(), aws.StringValue(container.State))
			return diag.Errorf("Lightsail Container Service deployment failed")
		}

		// disable the container service if needed
		if newIsDisabled {
			updateContainerServiceInput := lightsail.UpdateContainerServiceInput{
				ServiceName: serviceName,
				IsDisabled:  aws.Bool(true),
			}

			if _, err := conn.UpdateContainerService(&updateContainerServiceInput); err != nil {
				log.Printf("[ERROR] Lightsail Container Service (%s) update failed: %s", d.Id(), err)
				return diag.FromErr(err)
			}

			stateChangeConf := &resource.StateChangeConf{
				Pending:    []string{lightsail.ContainerServiceStateUpdating},
				Target:     []string{lightsail.ContainerServiceStateDisabled},
				Refresh:    lightsailContainerServiceRefreshFunc(serviceName, meta),
				Timeout:    25 * time.Minute,
				Delay:      5 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil {
				log.Printf("[ERROR] Lightsail Container Service (%s) error waiting for container service to disable: %s", d.Id(), err)
				return diag.FromErr(err)
			}

			log.Printf("[INFO] Lightsail Container Service (%s) disable successful", d.Id())
		}
	} else {
		// fields handled by UpdateContainerService, can update no matter container service disabled or not
		publicDomainNames, publicDomainNamesChanged := lightsailContainerServicePublicDomainNamesChanged(d)
		if d.HasChanges("is_disabled", "power", "scale") || publicDomainNamesChanged {
			// first determine Target, which is used by StateChangeConf later
			var targetState string
			if newIsDisabled {
				targetState = lightsail.ContainerServiceStateDisabled
			} else {
				resp, err := conn.GetContainerServices(&lightsail.GetContainerServicesInput{
					ServiceName: serviceName,
				})

				if err != nil {
					log.Printf("[ERROR] Lightsail Container Service (%s) update failed: %s", d.Id(), err)
					return diag.FromErr(err)
				}

				// index 0 because we only queried for 1 container service
				container := resp.ContainerServices[0]

				if container.CurrentDeployment != nil {
					targetState = lightsail.ContainerServiceStateRunning
				} else {
					targetState = lightsail.ContainerServiceStateReady
				}
			}

			updateContainerServiceInput := lightsail.UpdateContainerServiceInput{
				ServiceName:       serviceName,
				IsDisabled:        aws.Bool(newIsDisabled),
				Power:             aws.String(d.Get("power").(string)),
				PublicDomainNames: publicDomainNames,
				Scale:             aws.Int64(int64(d.Get("scale").(int))),
			}

			if _, err := conn.UpdateContainerService(&updateContainerServiceInput); err != nil {
				log.Printf("[ERROR] Lightsail Container Service (%s) update failed: %s", d.Id(), err)
				return diag.FromErr(err)
			}

			stateChangeConf := &resource.StateChangeConf{
				Pending:    []string{lightsail.ContainerServiceStateUpdating},
				Target:     []string{targetState},
				Refresh:    lightsailContainerServiceRefreshFunc(serviceName, meta),
				Timeout:    25 * time.Minute,
				Delay:      5 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil {
				log.Printf("[ERROR] Lightsail Container Service (%s) error waiting for container service to update: %s", d.Id(), err)
				return diag.FromErr(err)
			}
		}
	}

	log.Printf("[INFO] Lightsail Container Service (%s) update successful", d.Id())
	return resourceContainerServiceRead(ctx, d, meta)
}

func resourceContainerServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn
	serviceName := aws.String(d.Id())

	deleteContainerServiceInput := lightsail.DeleteContainerServiceInput{
		ServiceName: serviceName,
	}

	if _, err := conn.DeleteContainerService(&deleteContainerServiceInput); err != nil {
		log.Printf("[ERROR] Lightsail Container Service (%s) delete failed: %s", d.Id(), err)
		return diag.FromErr(err)
	}

	return nil
}

func expandLightsailContainerServiceDeployment(rawDeployment []interface{}) *lightsail.ContainerServiceDeploymentRequest {
	if len(rawDeployment) == 0 {
		return nil
	}

	deployment := lightsail.ContainerServiceDeploymentRequest{}

	for _, rd := range rawDeployment {
		rdMap := rd.(map[string]interface{})

		containers := rdMap["container"].(*schema.Set).List()
		if len(containers) > 0 {
			deployment.Containers = expandLightsailContainerServiceDeploymentContainers(containers)
		}

		publicEndpoint := rdMap["public_endpoint"].([]interface{})
		if len(publicEndpoint) > 0 {
			deployment.PublicEndpoint = expandLightsailContainerServiceDeploymentPublicEndpoint(publicEndpoint)
		}
	}

	return &deployment
}

func expandLightsailContainerServiceDeploymentContainers(rawContainers []interface{}) map[string]*lightsail.Container {
	if len(rawContainers) == 0 {
		return map[string]*lightsail.Container{}
	}

	result := make(map[string]*lightsail.Container)

	for _, rawContainer := range rawContainers {
		rawContainerMap := rawContainer.(map[string]interface{})

		containerName := rawContainerMap["container_name"].(string)
		// ignore empty-named container, which means a container is removed from .tf file
		// important to ignore this empty container because we don't need to delete an unwanted container,
		// besides, lightsail.CreateContainerServiceDeployment will throw InvalidInputException with an empty container name
		if containerName == "" {
			continue
		}

		container := lightsail.Container{
			Image: aws.String(rawContainerMap["image"].(string)),
		}

		var commands []*string
		// "command" is a []interface{} on top of []string, but Lightsail API needs a []*string
		for _, command := range rawContainerMap["command"].([]interface{}) {
			commands = append(commands, aws.String(command.(string)))
		}
		container.Command = commands

		environmentMap := make(map[string]*string)
		rawEnvironments := rawContainerMap["environment"].(*schema.Set).List()
		// rawEnvironment is a map[string]interface{} on top of map[string]string, but Lightsail API needs a map[string]*string
		for _, rawEnvironment := range rawEnvironments {
			rawEnvironmentMap := rawEnvironment.(map[string]interface{})
			environmentMap[rawEnvironmentMap["key"].(string)] = aws.String(rawEnvironmentMap["value"].(string))
		}
		container.Environment = environmentMap

		portsMap := make(map[string]*string)
		rawPorts := rawContainerMap["port"].(*schema.Set).List()
		// rawPort is a map[string]interface{} on top of map[string]string, but Lightsail API needs a map[string]*string
		for _, rawPort := range rawPorts {
			rawPortMap := rawPort.(map[string]interface{})
			portNumber := strconv.Itoa(rawPortMap["port_number"].(int))
			portsMap[portNumber] = aws.String(rawPortMap["protocol"].(string))
		}
		container.Ports = portsMap

		result[containerName] = &container
	}

	return result
}

func expandLightsailContainerServiceDeploymentPublicEndpoint(rawEndpoint []interface{}) *lightsail.EndpointRequest {
	if len(rawEndpoint) == 0 {
		return nil
	}

	endpoint := lightsail.EndpointRequest{}

	for _, re := range rawEndpoint {
		reMap := re.(map[string]interface{})

		endpoint.ContainerName = aws.String(reMap["container_name"].(string))

		endpoint.ContainerPort = aws.Int64(int64(reMap["container_port"].(int)))

		healthCheck := reMap["health_check"].([]interface{})
		if len(healthCheck) > 0 {
			endpoint.HealthCheck = expandLightsailContainerServiceDeploymentPublicEndpointHealthCheck(healthCheck)
		}
	}

	return &endpoint
}

func expandLightsailContainerServiceDeploymentPublicEndpointHealthCheck(rawHealthCheck []interface{}) *lightsail.ContainerServiceHealthCheckConfig {
	if len(rawHealthCheck) == 0 {
		return nil
	}

	healthCheck := lightsail.ContainerServiceHealthCheckConfig{}

	for _, rhc := range rawHealthCheck {
		rhcMap := rhc.(map[string]interface{})

		healthCheck.HealthyThreshold = aws.Int64(int64(rhcMap["healthy_threshold"].(int)))
		healthCheck.IntervalSeconds = aws.Int64(int64(rhcMap["interval_seconds"].(int)))
		healthCheck.Path = aws.String(rhcMap["path"].(string))
		healthCheck.SuccessCodes = aws.String(rhcMap["success_codes"].(string))
		healthCheck.TimeoutSeconds = aws.Int64(int64(rhcMap["timeout_seconds"].(int)))
		healthCheck.UnhealthyThreshold = aws.Int64(int64(rhcMap["unhealthy_threshold"].(int)))
	}

	return &healthCheck
}

func expandLightsailContainerServicePublicDomainNames(rawPublicDomainNames []interface{}) map[string][]*string {
	if len(rawPublicDomainNames) == 0 {
		return nil
	}

	resultMap := make(map[string][]*string)

	for _, rpdn := range rawPublicDomainNames {
		rpdnMap := rpdn.(map[string]interface{})

		rawCertificates := rpdnMap["certificate"].(*schema.Set).List()

		for _, rc := range rawCertificates {
			rcMap := rc.(map[string]interface{})

			var domainNames []*string
			for _, rawDomainName := range rcMap["domain_names"].([]interface{}) {
				domainNames = append(domainNames, aws.String(rawDomainName.(string)))
			}

			certificateName := rcMap["certificate_name"].(string)

			resultMap[certificateName] = domainNames
		}
	}

	return resultMap
}

func flattenLightsailContainerServiceDeployment(deployment *lightsail.ContainerServiceDeployment) []interface{} {
	if deployment == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"container":       flattenLightsailContainerServiceDeploymentContainers(deployment.Containers),
			"public_endpoint": flattenLightsailContainerServiceDeploymentPublicEndpoint(deployment.PublicEndpoint),
			"state":           aws.StringValue(deployment.State),
			"version":         int(aws.Int64Value(deployment.Version)),
		},
	}
}

func flattenLightsailContainerServiceDeploymentContainers(containers map[string]*lightsail.Container) []interface{} {
	if containers == nil {
		return []interface{}{}
	}

	var rawContainers []interface{}
	for containerName, container := range containers {
		rawContainer := map[string]interface{}{
			"container_name": containerName,
			"image":          aws.StringValue(container.Image),
			"command":        aws.StringValueSlice(container.Command),
			"environment":    flattenLightsailContainerServiceDeploymentEnvironment(container.Environment),
			"port":           flattenLightsailContainerServiceDeploymentPort(container.Ports),
		}

		rawContainers = append(rawContainers, rawContainer)
	}

	return rawContainers
}

func flattenLightsailContainerServiceDeploymentEnvironment(environment map[string]*string) []interface{} {
	if len(environment) == 0 {
		return []interface{}{}
	}

	var rawEnvironment []interface{}
	for key, value := range environment {
		rawEnvironment = append(rawEnvironment, map[string]interface{}{
			"key":   key,
			"value": aws.StringValue(value),
		})
	}
	return rawEnvironment
}

func flattenLightsailContainerServiceDeploymentPort(port map[string]*string) []interface{} {
	if len(port) == 0 {
		return []interface{}{}
	}

	var rawPorts []interface{}
	for portNumber, protocol := range port {
		portNumber, err := strconv.Atoi(portNumber)
		if err != nil {
			return []interface{}{}
		}
		rawPorts = append(rawPorts, map[string]interface{}{
			"port_number": portNumber,
			"protocol":    aws.StringValue(protocol),
		})
	}
	return rawPorts
}

func flattenLightsailContainerServiceDeploymentPublicEndpoint(endpoint *lightsail.ContainerServiceEndpoint) []interface{} {
	if endpoint == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"container_name": aws.StringValue(endpoint.ContainerName),
			"container_port": int(aws.Int64Value(endpoint.ContainerPort)),
			"health_check":   flattenLightsailContainerServiceDeploymentPublicEndpointHealthCheck(endpoint.HealthCheck),
		},
	}
}

func flattenLightsailContainerServiceDeploymentPublicEndpointHealthCheck(healthCheck *lightsail.ContainerServiceHealthCheckConfig) []interface{} {
	if healthCheck == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"healthy_threshold":   int(aws.Int64Value(healthCheck.HealthyThreshold)),
			"interval_seconds":    int(aws.Int64Value(healthCheck.IntervalSeconds)),
			"path":                aws.StringValue(healthCheck.Path),
			"success_codes":       aws.StringValue(healthCheck.SuccessCodes),
			"timeout_seconds":     int(aws.Int64Value(healthCheck.TimeoutSeconds)),
			"unhealthy_threshold": int(aws.Int64Value(healthCheck.UnhealthyThreshold)),
		},
	}
}

func flattenLightsailContainerServicePublicDomainNames(domainNames map[string][]*string) []interface{} {
	if domainNames == nil {
		return []interface{}{}
	}

	var rawCertificates []interface{}

	for certName, domains := range domainNames {
		rawCertificate := map[string]interface{}{
			"certificate_name": certName,
			"domain_names":     aws.StringValueSlice(domains),
		}

		rawCertificates = append(rawCertificates, rawCertificate)
	}

	return []interface{}{
		map[string]interface{}{
			"certificate": rawCertificates,
		},
	}
}

// call GetContainerServices to check the current state of the container service
func lightsailContainerServiceRefreshFunc(serviceName *string, m interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := m.(*conns.AWSClient).LightsailConn
		log.Printf("[DEBUG] Checking Lightsail Container Service state changes")
		resp, err := conn.GetContainerServices(
			&lightsail.GetContainerServicesInput{
				ServiceName: serviceName,
			},
		)
		if err != nil {
			return nil, "", err
		}
		return resp.ContainerServices[0], aws.StringValue(resp.ContainerServices[0].State), nil
	}
}

func lightsailContainerServiceDeploymentChanged(d *schema.ResourceData) (*lightsail.ContainerServiceDeploymentRequest, bool) {
	o, n := d.GetChange("deployment")
	oldDeployment := expandLightsailContainerServiceDeployment(o.([]interface{}))
	newDeployment := expandLightsailContainerServiceDeployment(n.([]interface{}))
	return newDeployment, !reflect.DeepEqual(oldDeployment, newDeployment)
}

func lightsailContainerServicePublicDomainNamesChanged(d *schema.ResourceData) (map[string][]*string, bool) {
	o, n := d.GetChange("public_domain_names")
	oldPublicDomainNames := expandLightsailContainerServicePublicDomainNames(o.([]interface{}))
	newPublicDomainNames := expandLightsailContainerServicePublicDomainNames(n.([]interface{}))

	changed := !reflect.DeepEqual(oldPublicDomainNames, newPublicDomainNames)
	if changed {
		if newPublicDomainNames == nil {
			newPublicDomainNames = map[string][]*string{}
		}

		// if the change is to detach a certificate, in .tf, a certificate block is removed
		// however, an empty []*string entry must be added to tell Lightsail that we want none of the domain names
		// under the certificate, effectively detaching the certificate
		for certificateName := range oldPublicDomainNames {
			if _, ok := newPublicDomainNames[certificateName]; !ok {
				newPublicDomainNames[certificateName] = []*string{}
			}
		}
	}

	return newPublicDomainNames, changed
}
