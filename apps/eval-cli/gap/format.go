package gap

import (
	"os"
	"strings"
)

// FormatToExt maps MIME format strings to file extensions.
var FormatToExt = map[string]string{
	"text/html":              ".html",
	"text/x-python":          ".py",
	"application/javascript": ".js",
	"text/typescript":         ".ts",
	"application/json":       ".json",
	"text/x-yaml":            ".yaml",
	"text/x-toml":            ".toml",
	"text/x-rust":            ".rs",
	"text/x-go":              ".go",
	"text/css":               ".css",
	"text/x-shellscript":     ".sh",
	"text/markdown":          ".md",
	"image/svg+xml":          ".svg",
	"application/xml":        ".xml",
	"text/x-java":            ".java",
	"text/x-ruby":            ".rb",
	"application/sql":        ".sql",
}

// ParseExperimentFormat reads a README.md and extracts the **Format:** value.
// Returns (format, ext). Defaults to ("text/html", ".html") if not found.
func ParseExperimentFormat(readmePath string) (string, string) {
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return "text/html", ".html"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "**Format:**") {
			parts := strings.SplitN(line, "**Format:**", 2)
			if len(parts) < 2 {
				continue
			}
			fmt := strings.TrimSpace(strings.SplitN(parts[1], "|", 2)[0])
			if ext, ok := FormatToExt[fmt]; ok {
				return fmt, ext
			}
		}
	}
	return "text/html", ".html"
}
