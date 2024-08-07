package utils

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func ParseTemplateVariables(
	content string,
	expectedVariables []string,
	templateVariables map[string]string,
) (string, error) {
	for _, expectedVariable := range expectedVariables {
		value, ok := templateVariables[expectedVariable]

		if !ok {
			return "", status.Errorf(
				codes.InvalidArgument,
				"missing template variable {%s}",
				expectedVariable,
			)
		}

		content = strings.ReplaceAll(content, fmt.Sprintf("{%s}", expectedVariable), value)
	}

	return content, nil
}
