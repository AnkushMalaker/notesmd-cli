package actions_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Yakitrak/notesmd-cli/pkg/actions"
	"github.com/Yakitrak/notesmd-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestApplyTemplateVariables(t *testing.T) {
	t.Run("Substitutes title, date and time with defaults", func(t *testing.T) {
		// Arrange
		content := "# {{title}}\ncreated: {{date}}\nat {{time}}"
		// Act
		result := actions.ApplyTemplateVariables(content, "People/Alice.md", obsidian.TemplatesConfig{})
		// Assert
		today := time.Now().Format("2006-01-02")
		assert.Contains(t, result, "# Alice")
		assert.Contains(t, result, "created: "+today)
		assert.NotContains(t, result, "{{")
	})

	t.Run("Honors inline date format override", func(t *testing.T) {
		// Arrange
		content := "{{date:YYYY}}"
		// Act
		result := actions.ApplyTemplateVariables(content, "note", obsidian.TemplatesConfig{})
		// Assert
		assert.Equal(t, time.Now().Format("2006"), result)
	})

	t.Run("Honors configured date format", func(t *testing.T) {
		// Arrange
		content := "{{date}}"
		cfg := obsidian.TemplatesConfig{DateFormat: "YYYY/MM/DD"}
		// Act
		result := actions.ApplyTemplateVariables(content, "note", cfg)
		// Assert
		assert.Equal(t, time.Now().Format("2006/01/02"), result)
	})

	t.Run("Leaves content without variables unchanged", func(t *testing.T) {
		// Arrange
		content := "plain content, no variables"
		// Act
		result := actions.ApplyTemplateVariables(content, "note", obsidian.TemplatesConfig{})
		// Assert
		assert.Equal(t, content, result)
	})
}

func TestResolveTemplate(t *testing.T) {
	t.Run("Resolves from configured Templates folder", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		obsDir := filepath.Join(tmpDir, ".obsidian")
		if err := os.MkdirAll(obsDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(obsDir, "templates.json"), []byte(`{"folder":"Templates"}`), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(tmpDir, "Templates"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "Templates", "Topic Template.md"), []byte("body {{title}}"), 0644); err != nil {
			t.Fatal(err)
		}
		// Act
		out, err := actions.ResolveTemplate(tmpDir, "Topic Template", "Topics/AI Safety")
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "body AI Safety", out)
	})

	t.Run("Falls back to vault-relative path when no templates folder", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "T.md"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		// Act
		out, err := actions.ResolveTemplate(tmpDir, "T", "note")
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "x", out)
	})

	t.Run("Errors when template is missing", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		// Act
		_, err := actions.ResolveTemplate(tmpDir, "Missing", "note")
		// Assert
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "not found"))
	})
}
