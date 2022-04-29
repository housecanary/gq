package ast

import "testing"

func TestStringRepresentation(t *testing.T) {
	sv := StringValue{V: "\r\n\b	\u0004\u0013\u0123\u1234"}
	expected := `"\r\n\b	\u0004\u0013ģሴ"`
	if sv.Representation() != expected {
		t.Fatalf("expected %s, got %s", expected, sv.Representation())
	}
}
