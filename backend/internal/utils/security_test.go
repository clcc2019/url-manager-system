package utils

import (
	"testing"
)

func TestSanitizeKubernetesName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"豆瓣", "unnamed"},
		{"test-project", "test-project"},
		{"Test Project", "test-project"},
		{"项目_测试", "unnamed"},
		{"my-project-123", "my-project-123"},
		{"", "unnamed"},
		{"   ", "unnamed"},
		{"123project", "123project"},
		{"project-", "project"},
		{"-project", "project"},
		{"very-long-project-name-that-exceeds-the-kubernetes-limit-of-sixty-three-characters", "very-long-project-name-that-exceeds-the-kubernetes-limit-of-six"},
		{"special@#$%chars", "special-chars"},
		{"中文项目名称", "unnamed"},
	}

	for _, test := range tests {
		result := SanitizeKubernetesName(test.input)
		if result != test.expected {
			t.Errorf("SanitizeKubernetesName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestSanitizeKubernetesLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"豆瓣", "unnamed"},
		{"test-project", "test-project"},
		{"Test_Project.v1", "Test_Project.v1"},
		{"", ""},
		{"   ", "unnamed"},
		{"123project", "123project"},
		{"project@#$%", "project"},
		{"very-long-label-value-that-exceeds-the-kubernetes-limit-of-sixty-three-characters", "very-long-label-value-that-exceeds-the-kubernetes-limit-of-sixt"},
		{"中文标签", "unnamed"},
	}

	for _, test := range tests {
		result := SanitizeKubernetesLabel(test.input)
		if result != test.expected {
			t.Errorf("SanitizeKubernetesLabel(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
