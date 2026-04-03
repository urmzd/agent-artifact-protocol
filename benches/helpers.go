package main

import "strings"

// sectionMarkerInstruction returns format-specific marker instructions.
func sectionMarkerInstruction(format string) string {
	switch {
	case strings.Contains(format, "html"), strings.Contains(format, "markdown"),
		strings.Contains(format, "svg"), strings.Contains(format, "xml"):
		return "Use AAP section markers: wrap each major block with `<!-- section:id -->` and `<!-- /section:id -->`."
	case strings.Contains(format, "javascript"), strings.Contains(format, "typescript"),
		strings.Contains(format, "rust"), strings.Contains(format, "go"),
		strings.Contains(format, "java"), strings.Contains(format, "css"):
		return "Use AAP section markers: wrap each major block with `// #region id` and `// #endregion id`."
	case strings.Contains(format, "python"), strings.Contains(format, "ruby"),
		strings.Contains(format, "sh"), strings.Contains(format, "yaml"):
		return "Use AAP section markers: wrap each major block with `# region id` and `# endregion id`."
	default:
		return "Use AAP section markers: wrap each major block with `<!-- section:id -->` and `<!-- /section:id -->`."
	}
}
