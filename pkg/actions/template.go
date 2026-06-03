package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Yakitrak/notesmd-cli/pkg/obsidian"
)

// templateVarPattern matches Obsidian core template variables: {{title}}, {{date}},
// {{time}}, and their format variants {{date:YYYY-MM-DD}} / {{time:HH:mm}}.
var templateVarPattern = regexp.MustCompile(`{{\s*(title|date|time)(?::([^}]*))?\s*}}`)

// ResolveTemplate reads the named template's content from the vault and substitutes
// the core template variables. The template is looked up in the Templates folder
// configured in .obsidian/templates.json (when the name has no explicit path),
// falling back to a vault-relative path. Returns an error if it cannot be read.
func ResolveTemplate(vaultPath, templateName, noteName string) (string, error) {
	cfg := obsidian.ReadTemplatesConfig(vaultPath)

	name := obsidian.AddMdSuffix(templateName)
	candidates := make([]string, 0, 2)
	if cfg.Folder != "" && !strings.Contains(templateName, "/") {
		candidates = append(candidates, filepath.Join(vaultPath, cfg.Folder, name))
	}
	candidates = append(candidates, filepath.Join(vaultPath, name))

	for _, c := range candidates {
		if data, err := os.ReadFile(c); err == nil {
			return ApplyTemplateVariables(string(data), noteName, cfg), nil
		}
	}
	return "", fmt.Errorf("template %q not found in vault (looked in %v)", templateName, candidates)
}

// ApplyTemplateVariables substitutes {{title}}, {{date}} and {{time}} (with optional
// inline Moment.js formats) in template content. Dates/times use the formats from
// the Templates config, defaulting to YYYY-MM-DD and HH:mm.
func ApplyTemplateVariables(content, noteName string, cfg obsidian.TemplatesConfig) string {
	now := time.Now()
	title := obsidian.RemoveMdSuffix(filepath.Base(noteName))

	defaultDate := cfg.DateFormat
	if defaultDate == "" {
		defaultDate = "YYYY-MM-DD"
	}
	defaultTime := cfg.TimeFormat
	if defaultTime == "" {
		defaultTime = "HH:mm"
	}

	return templateVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := templateVarPattern.FindStringSubmatch(match)
		variable, format := sub[1], sub[2]
		switch variable {
		case "title":
			return title
		case "date":
			if format == "" {
				format = defaultDate
			}
			return now.Format(obsidian.MomentToGoFormat(format))
		case "time":
			if format == "" {
				format = defaultTime
			}
			return now.Format(obsidian.MomentToGoFormat(format))
		}
		return match
	})
}
