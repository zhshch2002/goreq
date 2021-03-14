package goreq

import "testing"

func TestModifyLink(t *testing.T) {
	src := []string{"127.0.0.1", ":8080/query", "/query", "http://127.0.0.1", "https://127.0.0.1"}
	want := []string{"http://127.0.0.1", "http://localhost:8080/query", "http://localhost/query", "http://127.0.0.1", "https://127.0.0.1"}

	for k, v := range src {
		if want[k] != ModifyLink(v) {
			t.Errorf("got %s want %s\n", ModifyLink(v), want[k])
		}
	}
}
