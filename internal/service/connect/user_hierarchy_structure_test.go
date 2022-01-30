package connect_test

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccUserHierarchyStructureBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccUserHierarchyStructureBasicConfig(rName, levelOneName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }
  }
}
`, levelOneName))
}

func testAccUserHierarchyStructureBasicTwoLevelsConfig(rName, levelOneName, levelTwoName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }

    level_two {
      name = %[2]q
    }
  }
}
`, levelOneName, levelTwoName))
}

func testAccUserHierarchyStructureBasicThreeLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }

    level_two {
      name = %[2]q
    }

    level_three {
      name = %[3]q
    }
  }
}
`, levelOneName, levelTwoName, levelThreeName))
}

func testAccUserHierarchyStructureBasicFourLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName, levelFourName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }

    level_two {
      name = %[2]q
    }

    level_three {
      name = %[3]q
    }

    level_four {
      name = %[4]q
    }
  }
}
`, levelOneName, levelTwoName, levelThreeName, levelFourName))
}

func testAccUserHierarchyStructureBasicFiveLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName, levelFourName, levelFiveName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }

    level_two {
      name = %[2]q
    }

    level_three {
      name = %[3]q
    }

    level_four {
      name = %[4]q
    }

    level_five {
      name = %[5]q
    }
  }
}
`, levelOneName, levelTwoName, levelThreeName, levelFourName, levelFiveName))
}
