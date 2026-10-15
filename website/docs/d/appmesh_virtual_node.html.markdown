---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_node"
description: |-
    Terraform data source for managing an AWS App Mesh Virtual Node.
---

# Data Source: aws_appmesh_virtual_node

Terraform data source for managing an AWS App Mesh Virtual Node.

## Example Usage

```terraform
data "aws_appmesh_virtual_node" "test" {
  name      = "serviceBv1"
  mesh_name = "example-mesh"
}
```

## Argument Reference

This data source supports the following arguments:

* `mesh_name` - (Required) Name of the service mesh in which the virtual node exists.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.
* `name` - (Required) Name of the virtual node.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the virtual node.
* `created_date` - Creation date of the virtual node.
* `last_updated_date` - Last update date of the virtual node.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Virtual node specification. See [`spec` Block](#spec-block) for details.
* `tags` - Map of tags.

### `spec` Block

* `backend` - Backends to which the virtual node sends outbound traffic. See [`spec.backend` Block](#specbackend-block) for details.
* `backend_defaults` - Defaults for backends. See [`spec.backend_defaults` Block](#specbackend_defaults-block) for details.
* `listener` - Listeners from which the virtual node receives inbound traffic. See [`spec.listener` Block](#speclistener-block) for details.
* `logging` - Inbound and outbound access logging information for the virtual node. See [`spec.logging` Block](#speclogging-block) for details.
* `service_discovery` - Service discovery information for the virtual node. See [`spec.service_discovery` Block](#specservice_discovery-block) for details.

### `spec.backend` Block

* `virtual_service` - Virtual service used as a backend for a virtual node. See [`spec.backend.virtual_service` Block](#specbackendvirtual_service-block) for details.

### `spec.backend.virtual_service` Block

* `client_policy` - Client policy for the backend. See [`spec.backend.virtual_service.client_policy` Block](#specbackendvirtual_serviceclient_policy-block) for details.
* `virtual_service_name` - Name of the virtual service that is acting as a virtual node backend.

### `spec.backend.virtual_service.client_policy` Block

* `tls` - TLS client policy. See [`spec.backend.virtual_service.client_policy.tls` Block](#specbackendvirtual_serviceclient_policytls-block) for details.

### `spec.backend.virtual_service.client_policy.tls` Block

* `certificate` - Virtual node's client's TLS certificate. See [`spec.backend.virtual_service.client_policy.tls.certificate` Block](#specbackendvirtual_serviceclient_policytlscertificate-block) for details.
* `enforce` - Whether the policy is enforced.
* `ports` - One or more ports that the policy is enforced for.
* `validation` - TLS validation context. See [`spec.backend.virtual_service.client_policy.tls.validation` Block](#specbackendvirtual_serviceclient_policytlsvalidation-block) for details.

### `spec.backend.virtual_service.client_policy.tls.certificate` Block

* `file` - Local file certificate. See [`spec.backend.virtual_service.client_policy.tls.certificate.file` Block](#specbackendvirtual_serviceclient_policytlscertificatefile-block) for details.
* `sds` - [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate. See [`spec.backend.virtual_service.client_policy.tls.certificate.sds` Block](#specbackendvirtual_serviceclient_policytlscertificatesds-block) for details.

### `spec.backend.virtual_service.client_policy.tls.certificate.file` Block

* `certificate_chain` - Certificate chain for the certificate.
* `private_key` - Private key for a certificate stored on the file system of the mesh endpoint that the proxy is running on.

### `spec.backend.virtual_service.client_policy.tls.certificate.sds` Block

* `secret_name` - Name of the secret requested from the Secret Discovery Service provider representing TLS materials like a certificate or certificate chain.

### `spec.backend.virtual_service.client_policy.tls.validation` Block

* `subject_alternative_names` - SANs for a TLS validation context. See [`spec.backend.virtual_service.client_policy.tls.validation.subject_alternative_names` Block](#specbackendvirtual_serviceclient_policytlsvalidationsubject_alternative_names-block) for details.
* `trust` - TLS validation context trust. See [`spec.backend.virtual_service.client_policy.tls.validation.trust` Block](#specbackendvirtual_serviceclient_policytlsvalidationtrust-block) for details.

### `spec.backend.virtual_service.client_policy.tls.validation.subject_alternative_names` Block

* `match` - Criteria for determining a SAN's match. See [`spec.backend.virtual_service.client_policy.tls.validation.subject_alternative_names.match` Block](#specbackendvirtual_serviceclient_policytlsvalidationsubject_alternative_namesmatch-block) for details.

### `spec.backend.virtual_service.client_policy.tls.validation.subject_alternative_names.match` Block

* `exact` - Values sent must match the specified values exactly.

### `spec.backend.virtual_service.client_policy.tls.validation.trust` Block

* `acm` - TLS validation context trust for an AWS Certificate Manager (ACM) certificate. See [`spec.backend.virtual_service.client_policy.tls.validation.trust.acm` Block](#specbackendvirtual_serviceclient_policytlsvalidationtrustacm-block) for details.
* `file` - TLS validation context trust for a local file certificate. See [`spec.backend.virtual_service.client_policy.tls.validation.trust.file` Block](#specbackendvirtual_serviceclient_policytlsvalidationtrustfile-block) for details.
* `sds` - TLS validation context trust for a [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate. See [`spec.backend.virtual_service.client_policy.tls.validation.trust.sds` Block](#specbackendvirtual_serviceclient_policytlsvalidationtrustsds-block) for details.

### `spec.backend.virtual_service.client_policy.tls.validation.trust.acm` Block

* `certificate_authority_arns` - One or more ACM ARNs.

### `spec.backend.virtual_service.client_policy.tls.validation.trust.file` Block

* `certificate_chain` - Certificate trust chain for a certificate stored on the file system of the virtual node that the proxy is running on.

### `spec.backend.virtual_service.client_policy.tls.validation.trust.sds` Block

* `secret_name` - Name of the secret requested from the Secret Discovery Service provider representing TLS materials like a certificate or certificate chain.

### `spec.backend_defaults` Block

* `client_policy` - Default client policy for virtual service backends. See [`spec.backend_defaults.client_policy` Block](#specbackend_defaultsclient_policy-block) for details.

### `spec.backend_defaults.client_policy` Block

* `tls` - TLS client policy. See [`spec.backend_defaults.client_policy.tls` Block](#specbackend_defaultsclient_policytls-block) for details.

### `spec.backend_defaults.client_policy.tls` Block

* `certificate` - Virtual node's client's TLS certificate. See [`spec.backend_defaults.client_policy.tls.certificate` Block](#specbackend_defaultsclient_policytlscertificate-block) for details.
* `enforce` - Whether the policy is enforced.
* `ports` - One or more ports that the policy is enforced for.
* `validation` - TLS validation context. See [`spec.backend_defaults.client_policy.tls.validation` Block](#specbackend_defaultsclient_policytlsvalidation-block) for details.

### `spec.backend_defaults.client_policy.tls.certificate` Block

* `file` - Local file certificate. See [`spec.backend_defaults.client_policy.tls.certificate.file` Block](#specbackend_defaultsclient_policytlscertificatefile-block) for details.
* `sds` - [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate. See [`spec.backend_defaults.client_policy.tls.certificate.sds` Block](#specbackend_defaultsclient_policytlscertificatesds-block) for details.

### `spec.backend_defaults.client_policy.tls.certificate.file` Block

* `certificate_chain` - Certificate chain for the certificate.
* `private_key` - Private key for a certificate stored on the file system of the mesh endpoint that the proxy is running on.

### `spec.backend_defaults.client_policy.tls.certificate.sds` Block

* `secret_name` - Name of the secret requested from the Secret Discovery Service provider representing TLS materials like a certificate or certificate chain.

### `spec.backend_defaults.client_policy.tls.validation` Block

* `subject_alternative_names` - SANs for a TLS validation context. See [`spec.backend_defaults.client_policy.tls.validation.subject_alternative_names` Block](#specbackend_defaultsclient_policytlsvalidationsubject_alternative_names-block) for details.
* `trust` - TLS validation context trust. See [`spec.backend_defaults.client_policy.tls.validation.trust` Block](#specbackend_defaultsclient_policytlsvalidationtrust-block) for details.

### `spec.backend_defaults.client_policy.tls.validation.subject_alternative_names` Block

* `match` - Criteria for determining a SAN's match. See [`spec.backend_defaults.client_policy.tls.validation.subject_alternative_names.match` Block](#specbackend_defaultsclient_policytlsvalidationsubject_alternative_namesmatch-block) for details.

### `spec.backend_defaults.client_policy.tls.validation.subject_alternative_names.match` Block

* `exact` - Values sent must match the specified values exactly.

### `spec.backend_defaults.client_policy.tls.validation.trust` Block

* `acm` - TLS validation context trust for an AWS Certificate Manager (ACM) certificate. See [`spec.backend_defaults.client_policy.tls.validation.trust.acm` Block](#specbackend_defaultsclient_policytlsvalidationtrustacm-block) for details.
* `file` - TLS validation context trust for a local file certificate. See [`spec.backend_defaults.client_policy.tls.validation.trust.file` Block](#specbackend_defaultsclient_policytlsvalidationtrustfile-block) for details.
* `sds` - TLS validation context trust for a [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate. See [`spec.backend_defaults.client_policy.tls.validation.trust.sds` Block](#specbackend_defaultsclient_policytlsvalidationtrustsds-block) for details.

### `spec.backend_defaults.client_policy.tls.validation.trust.acm` Block

* `certificate_authority_arns` - One or more ACM ARNs.

### `spec.backend_defaults.client_policy.tls.validation.trust.file` Block

* `certificate_chain` - Certificate trust chain for a certificate stored on the file system of the virtual node that the proxy is running on.

### `spec.backend_defaults.client_policy.tls.validation.trust.sds` Block

* `secret_name` - Name of the secret requested from the Secret Discovery Service provider representing TLS materials like a certificate or certificate chain.

### `spec.listener` Block

* `connection_pool` - Connection pool information for the listener. See [`spec.listener.connection_pool` Block](#speclistenerconnection_pool-block) for details.
* `health_check` - Health check information for the listener. See [`spec.listener.health_check` Block](#speclistenerhealth_check-block) for details.
* `outlier_detection` - Outlier detection information for the listener. See [`spec.listener.outlier_detection` Block](#speclisteneroutlier_detection-block) for details.
* `port_mapping` - Port mapping information for the listener. See [`spec.listener.port_mapping` Block](#speclistenerport_mapping-block) for details.
* `timeout` - Timeouts for different protocols. See [`spec.listener.timeout` Block](#speclistenertimeout-block) for details.
* `tls` - TLS properties for the listener. See [`spec.listener.tls` Block](#speclistenertls-block) for details.

### `spec.listener.connection_pool` Block

* `grpc` - Connection pool information for gRPC listeners. See [`spec.listener.connection_pool.grpc` Block](#speclistenerconnection_poolgrpc-block) for details.
* `http` - Connection pool information for HTTP listeners. See [`spec.listener.connection_pool.http` Block](#speclistenerconnection_poolhttp-block) for details.
* `http2` - Connection pool information for HTTP2 listeners. See [`spec.listener.connection_pool.http2` Block](#speclistenerconnection_poolhttp2-block) for details.
* `tcp` - Connection pool information for TCP listeners. See [`spec.listener.connection_pool.tcp` Block](#speclistenerconnection_pooltcp-block) for details.

### `spec.listener.connection_pool.grpc` Block

* `max_requests` - Maximum number of inflight requests Envoy can concurrently support across hosts in upstream cluster.

### `spec.listener.connection_pool.http` Block

* `max_connections` - Maximum number of outbound TCP connections Envoy can establish concurrently with all hosts in upstream cluster.
* `max_pending_requests` - Number of overflowing requests after `max_connections` Envoy will queue to upstream cluster.

### `spec.listener.connection_pool.http2` Block

* `max_requests` - Maximum number of inflight requests Envoy can concurrently support across hosts in upstream cluster.

### `spec.listener.connection_pool.tcp` Block

* `max_connections` - Maximum number of outbound TCP connections Envoy can establish concurrently with all hosts in upstream cluster.

### `spec.listener.health_check` Block

* `healthy_threshold` - Number of consecutive successful health checks that must occur before declaring listener healthy.
* `interval_millis` - Time period in milliseconds between each health check execution.
* `path` - Destination path for the health check request.
* `port` - Destination port for the health check request.
* `protocol` - Protocol for the health check request.
* `timeout_millis` - Amount of time to wait when receiving a response from the health check, in milliseconds.
* `unhealthy_threshold` - Number of consecutive failed health checks that must occur before declaring a virtual node unhealthy.

### `spec.listener.outlier_detection` Block

* `base_ejection_duration` - Base amount of time for which a host is ejected. See [`spec.listener.outlier_detection.base_ejection_duration` Block](#speclisteneroutlier_detectionbase_ejection_duration-block) for details.
* `interval` - Time interval between ejection sweep analysis. See [`spec.listener.outlier_detection.interval` Block](#speclisteneroutlier_detectioninterval-block) for details.
* `max_ejection_percent` - Maximum percentage of hosts in load balancing pool for upstream service that can be ejected.
* `max_server_errors` - Number of consecutive `5xx` errors required for ejection.

### `spec.listener.outlier_detection.base_ejection_duration` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.outlier_detection.interval` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.port_mapping` Block

* `port` - Port used for the port mapping.
* `protocol` - Protocol used for the port mapping.

### `spec.listener.timeout` Block

* `grpc` - Timeouts for gRPC listeners. See [`spec.listener.timeout.grpc` Block](#speclistenertimeoutgrpc-block) for details.
* `http` - Timeouts for HTTP listeners. See [`spec.listener.timeout.http` Block](#speclistenertimeouthttp-block) for details.
* `http2` - Timeouts for HTTP2 listeners. See [`spec.listener.timeout.http2` Block](#speclistenertimeouthttp2-block) for details.
* `tcp` - Timeouts for TCP listeners. See [`spec.listener.timeout.tcp` Block](#speclistenertimeouttcp-block) for details.

### `spec.listener.timeout.grpc` Block

* `idle` - Idle timeout. See [`spec.listener.timeout.grpc.idle` Block](#speclistenertimeoutgrpcidle-block) for details.
* `per_request` - Per request timeout. See [`spec.listener.timeout.grpc.per_request` Block](#speclistenertimeoutgrpcper_request-block) for details.

### `spec.listener.timeout.grpc.idle` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.timeout.grpc.per_request` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.timeout.http` Block

* `idle` - Idle timeout. See [`spec.listener.timeout.http.idle` Block](#speclistenertimeouthttpidle-block) for details.
* `per_request` - Per request timeout. See [`spec.listener.timeout.http.per_request` Block](#speclistenertimeouthttpper_request-block) for details.

### `spec.listener.timeout.http.idle` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.timeout.http.per_request` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.timeout.http2` Block

* `idle` - Idle timeout. See [`spec.listener.timeout.http2.idle` Block](#speclistenertimeouthttp2idle-block) for details.
* `per_request` - Per request timeout. See [`spec.listener.timeout.http2.per_request` Block](#speclistenertimeouthttp2per_request-block) for details.

### `spec.listener.timeout.http2.idle` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.timeout.http2.per_request` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.timeout.tcp` Block

* `idle` - Idle timeout. See [`spec.listener.timeout.tcp.idle` Block](#speclistenertimeouttcpidle-block) for details.

### `spec.listener.timeout.tcp.idle` Block

* `unit` - Unit of time.
* `value` - Number of time units.

### `spec.listener.tls` Block

* `certificate` - Listener's TLS certificate. See [`spec.listener.tls.certificate` Block](#speclistenertlscertificate-block) for details.
* `mode` - Listener's TLS mode.
* `validation` - Listener's TLS validation context. See [`spec.listener.tls.validation` Block](#speclistenertlsvalidation-block) for details.

### `spec.listener.tls.certificate` Block

* `acm` - AWS Certificate Manager (ACM) certificate. See [`spec.listener.tls.certificate.acm` Block](#speclistenertlscertificateacm-block) for details.
* `file` - Local file certificate. See [`spec.listener.tls.certificate.file` Block](#speclistenertlscertificatefile-block) for details.
* `sds` - [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate. See [`spec.listener.tls.certificate.sds` Block](#speclistenertlscertificatesds-block) for details.

### `spec.listener.tls.certificate.acm` Block

* `certificate_arn` - ARN for the certificate.

### `spec.listener.tls.certificate.file` Block

* `certificate_chain` - Certificate chain for the certificate.
* `private_key` - Private key for a certificate stored on the file system of the virtual node that the proxy is running on.

### `spec.listener.tls.certificate.sds` Block

* `secret_name` - Name of the secret requested from the Secret Discovery Service provider representing TLS materials like a certificate or certificate chain.

### `spec.listener.tls.validation` Block

* `subject_alternative_names` - SANs for a TLS validation context. See [`spec.listener.tls.validation.subject_alternative_names` Block](#speclistenertlsvalidationsubject_alternative_names-block) for details.
* `trust` - TLS validation context trust. See [`spec.listener.tls.validation.trust` Block](#speclistenertlsvalidationtrust-block) for details.

### `spec.listener.tls.validation.subject_alternative_names` Block

* `match` - Criteria for determining a SAN's match. See [`spec.listener.tls.validation.subject_alternative_names.match` Block](#speclistenertlsvalidationsubject_alternative_namesmatch-block) for details.

### `spec.listener.tls.validation.subject_alternative_names.match` Block

* `exact` - Values sent must match the specified values exactly.

### `spec.listener.tls.validation.trust` Block

* `file` - TLS validation context trust for a local file certificate. See [`spec.listener.tls.validation.trust.file` Block](#speclistenertlsvalidationtrustfile-block) for details.
* `sds` - TLS validation context trust for a [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate. See [`spec.listener.tls.validation.trust.sds` Block](#speclistenertlsvalidationtrustsds-block) for details.

### `spec.listener.tls.validation.trust.file` Block

* `certificate_chain` - Certificate trust chain for a certificate stored on the file system of the mesh endpoint that the proxy is running on.

### `spec.listener.tls.validation.trust.sds` Block

* `secret_name` - Name of the secret for a virtual node's TLS Secret Discovery Service validation context trust.

### `spec.logging` Block

* `access_log` - Access log configuration for a virtual node. See [`spec.logging.access_log` Block](#specloggingaccess_log-block) for details.

### `spec.logging.access_log` Block

* `file` - File object to send virtual node access logs to. See [`spec.logging.access_log.file` Block](#specloggingaccess_logfile-block) for details.

### `spec.logging.access_log.file` Block

* `format` - Format for the logs. See [`spec.logging.access_log.file.format` Block](#specloggingaccess_logfileformat-block) for details.
* `path` - File path to write access logs to.

### `spec.logging.access_log.file.format` Block

* `json` - Logging format for JSON. See [`spec.logging.access_log.file.format.json` Block](#specloggingaccess_logfileformatjson-block) for details.
* `text` - Logging format for text.

### `spec.logging.access_log.file.format.json` Block

* `key` - Key for the JSON.
* `value` - Value for the JSON.

### `spec.service_discovery` Block

* `aws_cloud_map` - AWS Cloud Map information for the virtual node. See [`spec.service_discovery.aws_cloud_map` Block](#specservice_discoveryaws_cloud_map-block) for details.
* `dns` - DNS service name for the virtual node. See [`spec.service_discovery.dns` Block](#specservice_discoverydns-block) for details.

### `spec.service_discovery.aws_cloud_map` Block

* `attributes` - String map that contains attributes with values that you can use to filter instances by any custom attribute that you specified when you registered the instance.
* `namespace_name` - Name of the AWS Cloud Map namespace to use.
* `service_name` - Name of the AWS Cloud Map service to use.

### `spec.service_discovery.dns` Block

* `hostname` - DNS host name for your virtual node.
* `ip_preference` - Preferred IP version that this virtual node uses.
* `response_type` - DNS response type for the virtual node.
