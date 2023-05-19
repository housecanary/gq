package gen

import "testing"

func TestConverter(t *testing.T) {
	//s "AvmQuality,AvmQualityMethod" ~
	err := ConvertToTS(
		[]string{"ListingStatusData"},
		"/Users/mpoindexter/hc/property_graph_server/internal/pkg/basictypes",
		[]string{"/Users/mpoindexter/hc/property_graph_server/internal/pkg/basictypes"},
	)
	if err != nil {
		t.Fatal(err)
	}
}
