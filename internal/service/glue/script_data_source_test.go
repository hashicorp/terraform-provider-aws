// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueScriptDataSource_Language_python(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_glue_script.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScriptDataSourceConfig_python(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "python_script", regexache.MustCompile(`from awsglue\.job import Job`)),
				),
			},
		},
	})
}

func TestAccGlueScriptDataSource_Language_scala(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_glue_script.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScriptDataSourceConfig_scala(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "scala_code", regexache.MustCompile(`import com\.amazonaws\.services\.glue\.util\.Job`)),
				),
			},
		},
	})
}

func testAccScriptDataSourceConfig_python() string {
	return `
data "aws_glue_script" "test" {
  language = "PYTHON"

  dag_edge {
    source = "datasource0"
    target = "applymapping1"
  }

  dag_edge {
    source = "applymapping1"
    target = "selectfields2"
  }

  dag_edge {
    source = "selectfields2"
    target = "resolvechoice3"
  }

  dag_edge {
    source = "resolvechoice3"
    target = "datasink4"
  }

  dag_node {
    id        = "datasource0"
    node_type = "DataSource"

    args {
      name  = "database"
      value = "\"SourceDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"SourceTable\""
    }
  }

  dag_node {
    id        = "applymapping1"
    node_type = "ApplyMapping"

    args {
      name  = "mapping"
      value = "[(\"column1\", \"string\", \"column1\", \"string\")]"
    }
  }

  dag_node {
    id        = "selectfields2"
    node_type = "SelectFields"

    args {
      name  = "paths"
      value = "[\"column1\"]"
    }
  }

  dag_node {
    id        = "resolvechoice3"
    node_type = "ResolveChoice"

    args {
      name  = "choice"
      value = "\"MATCH_CATALOG\""
    }

    args {
      name  = "database"
      value = "\"DestinationDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"DestinationTable\""
    }
  }

  dag_node {
    id        = "datasink4"
    node_type = "DataSink"

    args {
      name  = "database"
      value = "\"DestinationDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"DestinationTable\""
    }
  }
}
`
}

func testAccScriptDataSourceConfig_scala() string {
	return `
data "aws_glue_script" "test" {
  language = "SCALA"

  dag_edge {
    source = "datasource0"
    target = "applymapping1"
  }

  dag_edge {
    source = "applymapping1"
    target = "selectfields2"
  }

  dag_edge {
    source = "selectfields2"
    target = "resolvechoice3"
  }

  dag_edge {
    source = "resolvechoice3"
    target = "datasink4"
  }

  dag_node {
    id        = "datasource0"
    node_type = "DataSource"

    args {
      name  = "database"
      value = "\"SourceDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"SourceTable\""
    }
  }

  dag_node {
    id        = "applymapping1"
    node_type = "ApplyMapping"

    args {
      name  = "mappings"
      value = "[(\"column1\", \"string\", \"column1\", \"string\")]"
    }
  }

  dag_node {
    id        = "selectfields2"
    node_type = "SelectFields"

    args {
      name  = "paths"
      value = "[\"column1\"]"
    }
  }

  dag_node {
    id        = "resolvechoice3"
    node_type = "ResolveChoice"

    args {
      name  = "choice"
      value = "\"MATCH_CATALOG\""
    }

    args {
      name  = "database"
      value = "\"DestinationDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"DestinationTable\""
    }
  }

  dag_node {
    id        = "datasink4"
    node_type = "DataSink"

    args {
      name  = "database"
      value = "\"DestinationDatabase\""
    }

    args {
      name  = "table_name"
      value = "\"DestinationTable\""
    }
  }
}
`
}
