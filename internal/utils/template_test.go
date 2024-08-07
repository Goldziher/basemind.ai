package utils_test

import (
	"github.com/basemind-ai/monorepo/internal/utils/testutils"
	"github.com/basemind-ai/monorepo/services/api-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMain(m *testing.M) {
	cleanup := testutils.CreateNamespaceTestDBModule("utils-test")
	defer cleanup()
	m.Run()
}

func TestUtils(t *testing.T) {
	t.Run("ParseTemplateVariables", func(t *testing.T) {
		t.Run("replaces all expected variables", func(t *testing.T) {
			content := "Hello {name}, your age is {age}. How are you {name}?"
			expectedVariables := []string{"name", "age"}
			templateVariables := map[string]string{"name": "John", "age": "30"}

			result, err := utils.ParseTemplateVariables(
				content,
				expectedVariables,
				templateVariables,
			)
			assert.NoError(t, err)

			expected := "Hello John, your age is 30. How are you John?"
			assert.Equal(t, expected, result)
		})
		t.Run(
			"returns the content string with no errors when there are no expected variables",
			func(t *testing.T) {
				content := "Hello {name}, your age is {age}. How are you {name}?"
				expectedVariables := make([]string, 0)
				templateVariables := map[string]string{"name": "John", "age": "30"}

				result, err := utils.ParseTemplateVariables(
					content,
					expectedVariables,
					templateVariables,
				)
				assert.NoError(t, err)

				assert.Equal(t, content, result)
			},
		)
		t.Run("returns an error when an expected variable is missing", func(t *testing.T) {
			content := "Hello {name}, your age is {age}."
			expectedVariables := []string{"name", "age"}
			templateVariables := map[string]string{"name": "John"}

			_, err := utils.ParseTemplateVariables(content, expectedVariables, templateVariables)
			assert.Error(t, err)

			expectedError := "missing template variable {age}"
			assert.Contains(t, err.Error(), expectedError)
		})
		t.Run("handles empty template variable", func(t *testing.T) {
			content := "Hello {name}, how are you?"
			expectedVariables := []string{"name"}
			templateVariables := map[string]string{"name": ""}

			result, err := utils.ParseTemplateVariables(
				content,
				expectedVariables,
				templateVariables,
			)
			assert.NoError(t, err)

			expected := "Hello , how are you?"
			assert.Equal(t, expected, result)
		})
	})
}
