// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"context"
	"fmt"
	"testing"

	acmpca_types "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/aws/aws-sdk-go/service/appmesh"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappmesh "github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVirtualNode_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_basic(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_basic(meshName, vnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappmesh.ResourceVirtualNode(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVirtualNode_backendClientPolicyACM(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	var ca acmpca_types.CertificateAuthority
	resourceName := "aws_appmesh_virtual_node.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"

	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			// We need to create and activate the CA before issuing a certificate.
			{
				Config: testAccVirtualNodeConfig_rootCA(domain),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(ctx, acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(ctx, &ca),
				),
			},
			{
				Config: testAccVirtualNodeConfig_backendClientPolicyACM(meshName, vnName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                                                                acctest.Ct1,
						"virtual_service.0.client_policy.#":                                                acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.#":                                          acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.certificate.#":                            acctest.Ct0,
						"virtual_service.0.client_policy.0.tls.0.enforce":                                  acctest.CtTrue,
						"virtual_service.0.client_policy.0.tls.0.ports.#":                                  acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.#":                             acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.subject_alternative_names.#": acctest.Ct0,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.#":                     acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.acm.#":               acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.file.#":              acctest.Ct0,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.sds.#":               acctest.Ct0,
						"virtual_service.0.virtual_service_name":                                           "servicea.simpleapp.local",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend.*.virtual_service.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "spec.0.backend.*.virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.acm.0.certificate_authority_arns.*", acmCAResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualNodeConfig_backendClientPolicyACM(meshName, vnName, domain),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(ctx, &ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVirtualNode_backendClientPolicyFile(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_backendClientPolicyFile(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                                                                     acctest.Ct1,
						"virtual_service.0.client_policy.#":                                                     acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.#":                                               acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.certificate.#":                                 acctest.Ct0,
						"virtual_service.0.client_policy.0.tls.0.enforce":                                       acctest.CtTrue,
						"virtual_service.0.client_policy.0.tls.0.ports.#":                                       acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.#":                                  acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.subject_alternative_names.#":      acctest.Ct0,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.#":                          acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.acm.#":                    acctest.Ct0,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.file.#":                   acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain": "/cert_chain.pem",
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.sds.#":                    acctest.Ct0,
						"virtual_service.0.virtual_service_name":                                                "servicea.simpleapp.local",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend.*.virtual_service.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVirtualNodeConfig_backendClientPolicyFileUpdated(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                                                                                acctest.Ct1,
						"virtual_service.0.client_policy.#":                                                                acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.#":                                                          acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.enforce":                                                  acctest.CtTrue,
						"virtual_service.0.client_policy.0.tls.0.ports.#":                                                  acctest.Ct2,
						"virtual_service.0.client_policy.0.tls.0.validation.#":                                             acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.subject_alternative_names.#":                 acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.#":         acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.#": acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.#":                                     acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.acm.#":                               acctest.Ct0,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.file.#":                              acctest.Ct1,
						"virtual_service.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain":            "/etc/ssl/certs/cert_chain.pem",
						"virtual_service.0.virtual_service_name":                                                           "servicea.simpleapp.local",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend.*.virtual_service.0.client_policy.0.tls.0.ports.*", "443"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend.*.virtual_service.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend.*.virtual_service.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "abc.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_backendDefaults(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_backendDefaults(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.enforce", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVirtualNodeConfig_backendDefaultsUpdated(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.enforce", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "443"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/etc/ssl/certs/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_backendDefaultsCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_backendDefaultsCertificate(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file.0.private_key", "tell-nobody"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.sds.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.enforce", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.ports.*", "8443"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "def.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds.0.secret_name", "restricted"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_cloudMapServiceDiscovery(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	nsResourceName := "aws_service_discovery_http_namespace.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// Avoid 'config is invalid: last character of "name" must be a letter' for aws_service_discovery_http_namespace.
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_cloudMapServiceDiscovery(meshName, vnName, rName, "Key1", "Value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.attributes.Key1", "Value1"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.namespace_name", nsResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.service_name", rName),
				),
			},
			{
				Config: testAccVirtualNodeConfig_cloudMapServiceDiscovery(meshName, vnName, rName, "Key1", "Value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.attributes.Key1", "Value2"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.namespace_name", nsResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.aws_cloud_map.0.service_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_listenerConnectionPool(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_listenerConnectionPool(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.grpc.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.tcp.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.tcp.0.max_connections", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.ip_preference", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.response_type", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVirtualNodeConfig_listenerConnectionPoolUpdated(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.grpc.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.0.max_connections", "8"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http.0.max_pending_requests", "16"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.http2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.0.tcp.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.ip_preference", "IPv4_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.response_type", "ENDPOINTS"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_listenerHealthChecks(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_listenerHealthChecks(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.healthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.interval_millis", "5000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.path", "/ping"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.protocol", "http2"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.timeout_millis", "2000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.unhealthy_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVirtualNodeConfig_listenerHealthChecksUpdated(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "servicec.simpleapp.local",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "serviced.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.healthy_threshold", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.interval_millis", "7000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.timeout_millis", "3000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.0.unhealthy_threshold", "9"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb1.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_listenerOutlierDetection(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_listenerOutlierDetection(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.base_ejection_duration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.base_ejection_duration.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.base_ejection_duration.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.interval.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.interval.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.interval.0.value", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.max_ejection_percent", "50"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.max_server_errors", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVirtualNodeConfig_listenerOutlierDetectionUpdated(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.base_ejection_duration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.base_ejection_duration.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.base_ejection_duration.0.value", "6"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.interval.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.interval.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.interval.0.value", "10000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.max_ejection_percent", "60"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.0.max_server_errors", "6"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_listenerTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_listenerTimeout(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.grpc.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.tcp.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.tcp.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.tcp.0.idle.0.unit", "ms"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.tcp.0.idle.0.value", "250000"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccVirtualNodeConfig_listenerTimeoutUpdated(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.grpc.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http.0.idle.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http.0.idle.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http.0.idle.0.value", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http.0.per_request.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http.0.per_request.0.unit", "s"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http.0.per_request.0.value", "5"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.http2.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.timeout.0.tcp.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVirtualNode_listenerTLS(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	var ca acmpca_types.CertificateAuthority
	resourceName := "aws_appmesh_virtual_node.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	acmCertificateResourceName := "aws_acm_certificate.test"

	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_listenerTLSFile(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.0.private_key", "/key.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			// We need to create and activate the CA before issuing a certificate.
			{
				Config: testAccVirtualNodeConfig_rootCA(domain),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(ctx, acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(ctx, &ca),
				),
			},
			{
				Config: testAccVirtualNodeConfig_listenerTLSACM(meshName, vnName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.0.certificate_arn", acmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "STRICT"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualNodeConfig_listenerTLSACM(meshName, vnName, domain),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(ctx, &ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVirtualNode_listenerValidation(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_listenerValidation(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.0.secret_name", "very-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "abc.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "xyz.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualNodeConfig_listenerValidationUpdated(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.0.secret_name", "top-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.0.secret_name", "confidential"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccVirtualNode_multiListenerValidation(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_multiListenerValidation(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.0.secret_name", "very-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "abc.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "xyz.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.0.sds.0.secret_name", "very-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.subject_alternative_names.0.match.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.subject_alternative_names.0.match.0.exact.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "abc.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.subject_alternative_names.0.match.0.exact.*", "xyz.example.com"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.0.file.0.certificate_chain", "/cert_chain.pem"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.0.sds.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualNodeConfig_multiListenerValidationUpdated(meshName, vnName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "spec.0.backend.*", map[string]string{
						"virtual_service.#":                      acctest.Ct1,
						"virtual_service.0.client_policy.#":      acctest.Ct0,
						"virtual_service.0.virtual_service_name": "servicea.simpleapp.local",
					}),
					resource.TestCheckResourceAttr(resourceName, "spec.0.backend_defaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.certificate.0.sds.0.secret_name", "top-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.subject_alternative_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.0.tls.0.validation.0.trust.0.sds.0.secret_name", "confidential"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.port_mapping.0.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.certificate.0.sds.0.secret_name", "top-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.subject_alternative_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.1.tls.0.validation.0.trust.0.sds.0.secret_name", "confidential"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.connection_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.health_check.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.outlier_detection.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.0.port", "8082"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.port_mapping.0.protocol", "grpc"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.certificate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.certificate.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.certificate.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.certificate.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.certificate.0.sds.0.secret_name", "top-secret"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.validation.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.validation.0.subject_alternative_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.validation.0.trust.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.validation.0.trust.0.acm.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.validation.0.trust.0.file.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.validation.0.trust.0.sds.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.listener.2.tls.0.validation.0.trust.0.sds.0.secret_name", "confidential"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.dns.0.hostname", "serviceb.simpleapp.local"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "appmesh", fmt.Sprintf("mesh/%s/virtualNode/%s", meshName, vnName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccVirtualNode_logging(t *testing.T) {
	ctx := acctest.Context(t)
	var vn appmesh.VirtualNodeData
	resourceName := "aws_appmesh_virtual_node.test"
	meshName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vnName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNodeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNodeConfig_logging(meshName, vnName, "/dev/stdout"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.format.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/dev/stdout"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccVirtualNodeImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVirtualNodeConfig_logging(meshName, vnName, "/tmp/access.log"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.format.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/tmp/access.log"),
				),
			},
			{
				Config: testAccVirtualNodeConfig_loggingWithFormat(meshName, vnName, "/tmp/access.log"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVirtualNodeExists(ctx, resourceName, &vn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, vnName),
					resource.TestCheckResourceAttr(resourceName, "mesh_name", meshName),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.format.0.json.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.format.0.json.0.key", "k1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.format.0.json.0.value", "v1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.format.0.text", ""),
					resource.TestCheckResourceAttr(resourceName, "spec.0.logging.0.access_log.0.file.0.path", "/tmp/access.log"),
				),
			},
		},
	})
}

func testAccCheckVirtualNodeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appmesh_virtual_node" {
				continue
			}

			_, err := tfappmesh.FindVirtualNodeByThreePartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Mesh Virtual Node %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVirtualNodeExists(ctx context.Context, n string, v *appmesh.VirtualNodeData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh Virtual Node ID is set")
		}

		output, err := tfappmesh.FindVirtualNodeByThreePartKey(ctx, conn, rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes["mesh_owner"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVirtualNodeImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["mesh_name"], rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccVirtualNodeConfig_mesh(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}
`, rName)
}

func testAccVirtualNodeConfig_rootCA(domain string) string {
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

func testAccVirtualNodeConfigPrivateCert(domain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = "test.%[1]s"
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}
`, domain)
}

func testAccVirtualNodeConfig_basic(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {}
}
`, vnName))
}

func testAccVirtualNodeConfig_backendDefaults(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
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
`, vnName))
}

func testAccVirtualNodeConfig_backendDefaultsUpdated(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
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
`, vnName))
}

func testAccVirtualNodeConfig_backendDefaultsCertificate(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend_defaults {
      client_policy {
        tls {
          certificate {
            file {
              certificate_chain = "/cert_chain.pem"
              private_key       = "tell-nobody"
            }
          }

          ports = [8443]

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
`, vnName))
}

func testAccVirtualNodeConfig_backendClientPolicyACM(meshName, vnName, domain string) string {
	return acctest.ConfigCompose(
		testAccVirtualNodeConfig_rootCA(domain),
		testAccVirtualNodeConfigPrivateCert(domain),
		testAccVirtualNodeConfig_mesh(meshName),
		fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"

        client_policy {
          tls {
            ports = [8443]

            validation {
              trust {
                acm {
                  certificate_authority_arns = [aws_acmpca_certificate_authority.test.arn]
                }
              }
            }
          }
        }
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_backendClientPolicyFile(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"

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

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_backendClientPolicyFileUpdated(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"

        client_policy {
          tls {
            ports = [443, 8443]

            validation {
              subject_alternative_names {
                match {
                  exact = ["abc.example.com"]
                }
              }

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

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_cloudMapServiceDiscovery(meshName, vnName, rName, attrKey, attrValue string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[2]q
}

resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      aws_cloud_map {
        attributes = {
          %[3]s = %[4]q
        }

        service_name   = %[2]q
        namespace_name = aws_service_discovery_http_namespace.test.name
      }
    }
  }
}
`, vnName, rName, attrKey, attrValue))
}

func testAccVirtualNodeConfig_listenerConnectionPool(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "tcp"
      }

      connection_pool {
        tcp {
          max_connections = 4
        }
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerConnectionPoolUpdated(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      connection_pool {
        http {
          max_connections      = 8
          max_pending_requests = 16
        }
      }
    }

    service_discovery {
      dns {
        hostname      = "serviceb.simpleapp.local"
        ip_preference = "IPv4_ONLY"
        response_type = "ENDPOINTS"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerHealthChecks(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

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

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerHealthChecksUpdated(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicec.simpleapp.local"
      }
    }

    backend {
      virtual_service {
        virtual_service_name = "serviced.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8081
        protocol = "http"
      }

      health_check {
        protocol            = "tcp"
        port                = 8081
        healthy_threshold   = 4
        unhealthy_threshold = 9
        timeout_millis      = 3000
        interval_millis     = 7000
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb1.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerOutlierDetection(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "tcp"
      }

      outlier_detection {
        base_ejection_duration {
          unit  = "ms"
          value = 250000
        }

        interval {
          unit  = "s"
          value = 10
        }

        max_ejection_percent = 50
        max_server_errors    = 5
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerOutlierDetectionUpdated(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      outlier_detection {
        base_ejection_duration {
          unit  = "s"
          value = 6
        }

        interval {
          unit  = "ms"
          value = 10000
        }

        max_ejection_percent = 60
        max_server_errors    = 6
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerTimeout(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "tcp"
      }

      timeout {
        tcp {
          idle {
            unit  = "ms"
            value = 250000
          }
        }
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerTimeoutUpdated(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      timeout {
        http {
          idle {
            unit  = "s"
            value = 10
          }

          per_request {
            unit  = "s"
            value = 5
          }
        }
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerTLSFile(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

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

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerTLSACM(meshName, vnName, domain string) string {
	return acctest.ConfigCompose(
		testAccVirtualNodeConfig_rootCA(domain),
		testAccVirtualNodeConfigPrivateCert(domain),
		testAccVirtualNodeConfig_mesh(meshName),
		fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

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

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerValidation(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

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

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_listenerValidationUpdated(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

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

        mode = "PERMISSIVE"

        validation {
          trust {
            sds {
              secret_name = "confidential"
            }
          }
        }
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_multiListenerValidation(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

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

    listener {
      port_mapping {
        port     = 8081
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

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_multiListenerValidationUpdated(meshName, vnName string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

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

        mode = "PERMISSIVE"

        validation {
          trust {
            sds {
              secret_name = "confidential"
            }
          }
        }
      }
    }

    listener {
      port_mapping {
        port     = 8081
        protocol = "http"
      }

      tls {
        certificate {
          sds {
            secret_name = "top-secret"
          }
        }

        mode = "PERMISSIVE"

        validation {
          trust {
            sds {
              secret_name = "confidential"
            }
          }
        }
      }
    }

    listener {
      port_mapping {
        port     = 8082
        protocol = "grpc"
      }

      tls {
        certificate {
          sds {
            secret_name = "top-secret"
          }
        }

        mode = "PERMISSIVE"

        validation {
          trust {
            sds {
              secret_name = "confidential"
            }
          }
        }
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName))
}

func testAccVirtualNodeConfig_logging(meshName, vnName, path string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    logging {
      access_log {
        file {
          path = %[2]q
        }
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName, path))
}

func testAccVirtualNodeConfig_loggingWithFormat(meshName, vnName, path string) string {
	return acctest.ConfigCompose(testAccVirtualNodeConfig_mesh(meshName), fmt.Sprintf(`
resource "aws_appmesh_virtual_node" "test" {
  name      = %[1]q
  mesh_name = aws_appmesh_mesh.test.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    logging {
      access_log {
        file {
          path = %[2]q

          format {
            json {
              key   = "k1"
              value = "v1"
            }
          }
        }
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
`, vnName, path))
}
