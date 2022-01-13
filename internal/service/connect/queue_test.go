func testAccQueueBaseConfig(rName string) string {
	// Use the aws_connect_hours_of_operation data source with the default "Basic Hours" that comes with connect instances.
	// Because if a resource is used, Terraform will not be able to delete it since queues do not have support for the delete api
	// yet but still references hours_of_operation_id. However, using the data source will result in the failure of the
	// disppears test (removed till delete api is available) for the connect instance (We test disappears on the Connect instance
	// instead of the queue since the queue does not support delete). The error is:
	// Step 1/1 error: Error running post-apply plan: exit status 1
	// Error: error finding Connect Hours of Operation Summary by name (Basic Hours): ResourceNotFoundException: Instance not found
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

data "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "Basic Hours"
}
`, rName)
}

func testAccQueueBasicConfig(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, label))
}
