package appmesh_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/appmesh"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"testing"
)

func TestAccAppMeshVirtualGatewayDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_basic(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_backendDefaults(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_backendDefaults(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.0", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.0"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_backendDefaultsUpdated(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_backendDefaultsUpdated(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.0", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.0"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.1", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_backendDefaultsCertificate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_backendDefaultsCertificate(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.0.certificate_chain", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.0.certificate_chain"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.0.private_key", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.0.private_key"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.exact.0", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.exact.0"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.0.secret_name", dataSourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.0.secret_name"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_listenerConnectionPool(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_listenerConnectionPool(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.connection_pool.0.grpc.0.max_requests", dataSourceName, "spec.0.listener.0.connection_pool.0.grpc.0.max_requests"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_listenerConnectionPoolUpdated(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_listenerConnectionPoolUpdated(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.connection_pool.0.http.0.max_connections", dataSourceName, "spec.0.listener.0.connection_pool.0.http.0.max_connections"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.connection_pool.0.http.0.max_pending_requests", dataSourceName, "spec.0.listener.0.connection_pool.0.http.0.max_pending_requests"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_listenerHealthChecks(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_listenerHealthChecks(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.health_check.0.protocol", dataSourceName, "spec.0.listener.0.health_check.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.health_check.0.path", dataSourceName, "spec.0.listener.0.health_check.0.path"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.health_check.0.healthy_threshold", dataSourceName, "spec.0.listener.0.health_check.0.healthy_threshold"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.health_check.0.unhealthy_threshold", dataSourceName, "spec.0.listener.0.health_check.0.unhealthy_threshold"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.health_check.0.timeout_millis", dataSourceName, "spec.0.listener.0.health_check.0.timeout_millis"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.health_check.0.interval_millis", dataSourceName, "spec.0.listener.0.health_check.0.interval_millis"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_listenerTLSACM(t *testing.T) {

	var ca acmpca.CertificateAuthority
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	domain := acctest.RandomDomainName()

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			// We need to create and activate the CA before issuing a certificate.
			{
				Config: testAccVirtualGatewayDataSourceConfig_rootCA(domain),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(&ca),
				),
			},
			{
				Config: testAccVirtualGatewayDataSourceConfig_listenerTLSACM(rName, vgName, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.0.certificate_arn", dataSourceName, "spec.0.listener.0.tls.0.certificate.0.acm.0.certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.mode", dataSourceName, "spec.0.listener.0.tls.0.mode"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_listenerTLSFile(t *testing.T) {

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_listenerTLSFile(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.certificate_chain", dataSourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.certificate_chain"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.private_key", dataSourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.private_key"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.mode", dataSourceName, "spec.0.listener.0.tls.0.mode"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_listenerValidation(t *testing.T) {

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_listenerValidation(rName, vgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.0.secret_name", dataSourceName, "spec.0.listener.0.tls.0.certificate.0.sds.0.secret_name"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.mode", dataSourceName, "spec.0.listener.0.tls.0.mode"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact", dataSourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.0.certificate_chain", dataSourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.0.certificate_chain"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_logging(t *testing.T) {

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())
	path := "/tmp/access.log"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_logging(rName, vgName, path),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.logging.access_log.0.file.0.path", dataSourceName, "spec.0.logging.access_log.0.file.0.path"),
				),
			},
		},
	})
}

func TestAccAppMeshVirtualGatewayDataSource_tags(t *testing.T) {

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_virtual_gateway.test"
	dataSourceName := "data.aws_appmesh_virtual_gateway.test"
	vgName := fmt.Sprintf("tf-acc-test-%d-mesh-local", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualGatewayDataSourceConfig_tags(rName, vgName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_date", dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "last_updated_date", dataSourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_owner", dataSourceName, "resource_owner"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.port", dataSourceName, "spec.0.listener.0.port_mapping.0.port"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.port_mapping.0.protocol", dataSourceName, "spec.0.listener.0.port_mapping.0.protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func testAccVirtualGatewayDataSourceConfig_basic(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_backendDefaults(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_backendDefaultsUpdated(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_backendDefaultsCertificate(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_listenerConnectionPool(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_listenerConnectionPoolUpdated(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_listenerHealthChecks(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_rootCA(domain string) string {
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

func testAccVirtualGatewayDataSourceConfig_listenerTLSACM(meshName, vgName, domain string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName, domain))
}

func testAccVirtualGatewayDataSourceConfig_listenerTLSFile(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_listenerValidation(meshName, vgName string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName)
}

func testAccVirtualGatewayDataSourceConfig_logging(meshName, vgName, path string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName, path)
}

func testAccVirtualGatewayDataSourceConfig_tags(meshName, vgName, tagKey, tagValue string) string {
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

data "aws_appmesh_virtual_gateway" "test" {
  name      = aws_appmesh_virtual_gateway.test.name
  mesh_name = aws_appmesh_mesh.test.name
}
`, meshName, vgName, tagKey, tagValue)
}
