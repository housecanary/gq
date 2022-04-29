package parser

import "testing"

func TestParseBlockStringIndent(t *testing.T) {
	parsed := parseBlockString(`"""
	  Hello,
	    World!
	
	  Yours,
	    GraphQL.
	"""`)

	expected := "Hello,\n  World!\n\nYours,\n  GraphQL."

	if parsed != expected {
		t.Fatalf("expected %s, got %s", expected, parsed)
	}

}

func TestParseBlockStringQuote(t *testing.T) {
	parsed := parseBlockString(`"""I have a \""" in me"""`)

	expected := `I have a """ in me`

	if parsed != expected {
		t.Fatalf("expected %s, got %s", expected, parsed)
	}

}

func TestParseStringEscapes(t *testing.T) {
	parsed := parseString(`"\"\\\/\b\f\n\r\t\u123A"`)

	expected := "\"\\/\b\f\n\r\t\u123A"

	if parsed != expected {
		t.Fatalf("expected %s, got %s", expected, parsed)
	}

}
