package domain

import "testing"

func TestExtractDomainPath(t *testing.T) {
	test1 := "example.com"
	if path, err := extractDomainPath(test1); err != nil {
		if path != "com/example" {
			t.Error("Domain path failed: ", test1, path)
		}
	}

	test2 := "test.example.com"
	if path, err := extractDomainPath(test2); err != nil {
		if path != "com/example/test" {
			t.Error("Domain path failed: ", test2, path)
		}
	}
}
