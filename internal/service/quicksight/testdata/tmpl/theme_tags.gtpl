resource "aws_quicksight_theme" "test" {
  theme_id = var.rName
  name     = var.rName

  base_theme_id = "MIDNIGHT"

  configuration {
    data_color_palette {
      colors = [
        "#FFFFFF",
        "#111111",
        "#222222",
        "#333333",
        "#444444",
        "#555555",
        "#666666",
        "#777777",
        "#888888",
        "#999999"
      ]
      empty_fill_color = "#FFFFFF"
      min_max_gradient = [
        "#FFFFFF",
        "#111111",
      ]
    }
  }
{{- template "tags" . }}
}
