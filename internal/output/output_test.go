// internal/output/output_test.go
package output

import (
	"bytes"
	"testing"
)

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test", "id": "123"}

	err := JSON(&buf, data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "{\n  \"id\": \"123\",\n  \"name\": \"test\"\n}\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestYAML(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test"}

	err := YAML(&buf, data)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("name: test")) {
		t.Errorf("expected YAML with 'name: test', got %q", buf.String())
	}
}
