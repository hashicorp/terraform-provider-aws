package appmesh_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappmesh "github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
)

func testAccGatewayRoute_basic(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vsResourceName := "aws_appmesh_virtual_service.test.0"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vsResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_disappears(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfappmesh.ResourceGatewayRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGatewayRoute_GRPCRoute(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_grpcRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				Config: testAccGatewayRouteConfig_grpcRouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.grpc_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.0.match.0.service_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_HTTPRoute(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_httpRoute(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				Config: testAccGatewayRouteConfig_httpRouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_HTTP2Route(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	vs1ResourceName := "aws_appmesh_virtual_service.test.0"
	vs2ResourceName := "aws_appmesh_virtual_service.test.1"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_http2Route(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs1ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				Config: testAccGatewayRouteConfig_http2RouteUpdated(meshName, vgName, grName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", grName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.grpc_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.http2_route.0.action.0.target.0.virtual_service.0.virtual_service_name", vs2ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http2_route.0.match.0.prefix", "/users"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.http_route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtual_gateway_name", vgName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "appmesh", fmt.Sprintf("mesh/%s/virtualGateway/%s/gatewayRoute/%s", meshName, vgName, grName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGatewayRoute_Tags(t *testing.T) {
	var v appmesh.GatewayRouteData
	resourceName := "aws_appmesh_gateway_route.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	grName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appmesh.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRouteConfig_tags1(meshName, vgName, grName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccGatewayRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGatewayRouteConfig_tags2(meshName, vgName, grName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGatewayRouteConfig_tags1(meshName, vgName, grName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccGatewayRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccCheckGatewayRouteDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appmesh_gateway_route" {
			continue
		}

		_, err := tfappmesh.FindGatewayRoute(conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes["name"], rs.Primary.Attributes["mesh_owner"])
		if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("App Mesh gateway route still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckGatewayRouteExists(name string, v *appmesh.GatewayRouteData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh gateway route ID is set")
		}

		out, err := tfappmesh.FindGatewayRoute(conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["virtual_gateway_name"], rs.Primary.Attributes["name"], rs.Primary.Attributes["mesh_owner"])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccGatewayRouteConfigBase(meshName, vgName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

resource "aws_appmesh_virtual_gateway" "test" {
  name      = %[2]q
  mesh_name = aws_appmesh_mesh.test.name

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_virtual_service" "test" {
  count = 2

  name      = "%[2]s-${count.index}"
  mesh_name = aws_appmesh_mesh.test.name

  spec {}
}
`, meshName, vgName)
}

func testAccGatewayRouteConfig_grpcRoute(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        service_name = "test1"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_grpcRouteUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    grpc_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        service_name = "test2"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRoute(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_httpRouteUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        prefix = "/users"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2Route(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_http2RouteUpdated(meshName, vgName, grName string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http2_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[1].name
          }
        }
      }

      match {
        prefix = "/users"
      }
    }
  }
}
`, grName))
}

func testAccGatewayRouteConfig_tags1(meshName, vgName, grName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, grName, tagKey1, tagValue1))
}

func testAccGatewayRouteConfig_tags2(meshName, vgName, grName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccGatewayRouteConfigBase(meshName, vgName), fmt.Sprintf(`
resource "aws_appmesh_gateway_route" "test" {
  name                 = %[1]q
  mesh_name            = aws_appmesh_mesh.test.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.test.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.test[0].name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, grName, tagKey1, tagValue1, tagKey2, tagValue2))
}
