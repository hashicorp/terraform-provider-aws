resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  {{- template "region" }}
  name = "tf-acc-test-custom-routing-accelerator"
}
