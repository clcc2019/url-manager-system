package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateImageName(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected bool
	}{
		{"Valid simple image", "nginx", true},
		{"Valid image with tag", "nginx:latest", true},
		{"Valid image with registry", "registry.example.com/nginx:1.21", true},
		{"Valid image with namespace", "myorg/nginx:latest", true},
		{"Invalid empty image", "", false},
		{"Invalid with spaces", "nginx latest", false},
		{"Invalid with uppercase", "NGINX:LATEST", false},
		{"Valid with digest", "nginx@sha256:abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateImageName(tt.image)
			assert.Equal(t, tt.expected, result, "Image: %s", tt.image)
		})
	}
}

func TestValidateResourceString(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		expected bool
	}{
		{"Valid CPU millicores", "100m", true},
		{"Valid CPU cores", "0.5", true},
		{"Valid memory bytes", "512Mi", true},
		{"Valid memory with unit", "1Gi", true},
		{"Invalid empty string", "", false},
		{"Invalid format", "abc", false},
		{"Valid large memory", "10Gi", true},
		{"Valid small CPU", "50m", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateResourceString(tt.resource)
			assert.Equal(t, tt.expected, result, "Resource: %s", tt.resource)
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal string", "hello world", "hello world"},
		{"String with spaces", "  hello world  ", "hello world"},
		{"String with null byte", "hello\x00world", "helloworld"},
		{"String with newlines", "hello\nworld\r", "helloworld"},
		{"String with tabs", "hello\tworld", "helloworld"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateEnvironmentVariableName(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		expected bool
	}{
		{"Valid simple name", "VAR_NAME", true},
		{"Valid with underscore", "VAR_NAME_2", true},
		{"Invalid starting with number", "2VAR_NAME", false},
		{"Invalid with dash", "VAR-NAME", false},
		{"Invalid with space", "VAR NAME", false},
		{"Valid single letter", "A", true},
		{"Invalid empty", "", false},
		{"Valid lowercase", "var_name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEnvironmentVariableName(tt.envName)
			assert.Equal(t, tt.expected, result, "EnvName: %s", tt.envName)
		})
	}
}

func TestIsValidProjectName(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		expected    bool
	}{
		{"Valid project name", "my-project", true},
		{"Valid with underscore", "my_project", true},
		{"Valid with numbers", "project123", true},
		{"Invalid empty", "", false},
		{"Invalid too long", string(make([]byte, 101)), false},
		{"Invalid with spaces", "my project", false},
		{"Invalid with special chars", "my@project", false},
		{"Valid single char", "a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidProjectName(tt.projectName)
			assert.Equal(t, tt.expected, result, "ProjectName: %s", tt.projectName)
		})
	}
}

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"Same strings", "hello", "hello", true},
		{"Different strings", "hello", "world", false},
		{"Empty strings", "", "", true},
		{"One empty", "hello", "", false},
		{"Case sensitive", "Hello", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SecureCompare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}