package appmesh_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappmesh "github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
)

func testAccVirtualGateway_basic(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_basic(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualGateway_disappears(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_basic(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfappmesh.ResourceVirtualGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVirtualGateway_BackendDefaults(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_backendDefaults(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.enforce", "true"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				Config: testAccVirtualGatewayConfig_backendDefaultsUpdated(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.enforce", "true"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "443"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/etc/ssl/certs/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualGateway_BackendDefaultsCertificate(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_backendDefaultsCertificate(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.0.private_key", "tell-nobody"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.sds.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.enforce", "true"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "def.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.0.secret_name", "restricted"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualGateway_ListenerConnectionPool(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_listenerConnectionPool(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.grpc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.grpc.0.max_requests", "4"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				Config: testAccVirtualGatewayConfig_listenerConnectionPoolUpdated(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.grpc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.0.max_connections", "8"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.0.max_pending_requests", "16"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http2.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualGateway_ListenerHealthChecks(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_listenerHealthChecks(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.interval_millis", "5000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.path", "/ping"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.protocol", "http2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.timeout_millis", "2000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.unhealthy_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				Config: testAccVirtualGatewayConfig_listenerHealthChecksUpdated(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.interval_millis", "7000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.path", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.timeout_millis", "3000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.unhealthy_threshold", "9"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualGateway_ListenerTLS(t *testing.T) {
	var v appmesh.VirtualGatewayData
	var ca acmpca.CertificateAuthority
	resourceName := "aws_appmesh_virtual_gateway.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	acmCertificateResourceName := "aws_acm_certificate.test"

	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_listenerTLSFile(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.private_key", "/key.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			// We need to create and activate the CA before issuing a certificate.
			{
				Config: testAccVirtualGatewayConfig_rootCA(domain),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(&ca),
				),
			},
			{
				Config: testAccVirtualGatewayConfig_listenerTLSACM(meshName, vgName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.0.certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "STRICT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualGatewayConfig_listenerTLSACM(meshName, vgName, domain),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(&ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVirtualGateway_ListenerValidation(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_listenerValidation(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.0.secret_name", "very-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "abc.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "xyz.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualGatewayConfig_listenerValidationUpdated(meshName, vgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.0.secret_name", "top-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "STRICT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.acm.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.0.secret_name", "confidential"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
		},
	})
}

func testAccVirtualGateway_Logging(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_logging(meshName, vgName, "/dev/stdout"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/dev/stdout"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				Config: testAccVirtualGatewayConfig_logging(meshName, vgName, "/tmp/access.log"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", vgName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/tmp/access.log"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s", meshName, vgName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualGateway_Tags(t *testing.T) {
	var v appmesh.VirtualGatewayData
	resourceName := "aws_appmesh_virtual_gateway.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVirtualGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayConfig_tags1(meshName, vgName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", meshName, vgName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualGatewayConfig_tags2(meshName, vgName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVirtualGatewayConfig_tags1(meshName, vgName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckVirtualGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_virtual_node" {
			continue
		}

		_, err := tfappmesh.FindVirtualGateway(conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["name"], rs.Primary.Attributes["mesh_owner"])
		if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("App Mesh virtual gateway still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVirtualGatewayExists(name string, v *appmesh.VirtualGatewayData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh virtual gateway ID is set")
		}

		out, err := tfappmesh.FindVirtualGateway(conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["name"], rs.Primary.Attributes["mesh_owner"])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccVirtualGatewayConfig_basic(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_backendDefaults(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    backend_defaults {
      client_policy {
        tls {
          ports = [8443]

          validation {
            trust {
              file {
                certificate_chain = "/cert_chain.pem"
              }
            }
          }
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_backendDefaultsUpdated(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    backend_defaults {
      client_policy {
        tls {
          ports = [443, 8443]

          validation {
            trust {
              file {
                certificate_chain = "/etc/ssl/certs/cert_chain.pem"
              }
            }
          }
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_backendDefaultsCertificate(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    backend_defaults {
      client_policy {
        tls {
          certificate {
            file {
              certificate_chain = "/cert_chain.pem"
              private_key       = "tell-nobody"
            }
          }

          validation {
            subject_alternative_names {
              match {
                exact = ["def.example.com"]
              }
            }

            trust {
              sds {
                secret_name = "restricted"
              }
            }
          }
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_listenerConnectionPool(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "grpc"
      }

      connection_pool {
        grpc {
          max_requests = 4
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_listenerConnectionPoolUpdated(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8081
        protocol = "http"
      }

      connection_pool {
        http {
          max_connections      = 8
          max_pending_requests = 16
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_listenerHealthChecks(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "grpc"
      }

      health_check {
        protocol            = "http2"
        path                = "/ping"
        healthy_threshold   = 3
        unhealthy_threshold = 5
        timeout_millis      = 2000
        interval_millis     = 5000
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_listenerHealthChecksUpdated(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8081
        protocol = "http2"
      }

      health_check {
        protocol            = "grpc"
        port                = 8081
        healthy_threshold   = 4
        unhealthy_threshold = 9
        timeout_millis      = 3000
        interval_millis     = 7000
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_rootCA(domain string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}
`, domain)
}

func testAccVirtualGatewayConfig_listenerTLSACM(meshName, vgName, domain string) string {
	return acctest.ConfigCompose(
		testAccVirtualGatewayConfig_rootCA(domain),
		fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_acm_certificate" "test" {
  domain_name               = "test.%[3]s"
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      tls {
        certificate {
          acm {
            certificate_arn = aws_acm_certificate.test.arn
          }
        }

        mode = "STRICT"
      }
    }
  }
}
`, meshName, vgName, domain))
}

func testAccVirtualGatewayConfig_listenerTLSFile(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      tls {
        certificate {
          file {
            certificate_chain = "/cert_chain.pem"
            private_key       = "/key.pem"
          }
        }

        mode = "PERMISSIVE"
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_listenerValidation(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      tls {
        certificate {
          sds {
            secret_name = "very-secret"
          }
        }

        mode = "PERMISSIVE"

        validation {
          subject_alternative_names {
            match {
              exact = ["abc.example.com", "xyz.example.com"]
            }
          }

          trust {
            file {
              certificate_chain = "/cert_chain.pem"
            }
          }
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_listenerValidationUpdated(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      tls {
        certificate {
          sds {
            secret_name = "top-secret"
          }
        }

        mode = "STRICT"

        validation {
          trust {
            sds {
              secret_name = "confidential"
            }
          }
        }
      }
    }
  }
}
`, meshName, vgName)
}

func testAccVirtualGatewayConfig_logging(meshName, vgName, path string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    logging {
      access_log {
        file {
          path = %[3]q
        }
      }
    }
  }
}
`, meshName, vgName, path)
}

func testAccVirtualGatewayConfig_tags1(meshName, vgName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, meshName, vgName, tagKey1, tagValue1)
}

func testAccVirtualGatewayConfig_tags2(meshName, vgName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, meshName, vgName, tagKey1, tagValue1, tagKey2, tagValue2)
}
