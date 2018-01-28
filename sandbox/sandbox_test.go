package sandbox

import "testing"


type version_port struct {
	version string
	port    int
}

func TestVersionToPort(t *testing.T) {
	t.Parallel()
	var versions []version_port = []version_port{
		{"", -1},            // FAIL: Empty string
		{"5.0.A", -1},       // FAIL: Hexadecimal number
		{"5.0.-9", -1},      // FAIL: Negative revision
		{"-5.0.9", -1},      // FAIL: Negative major version
		{"5.-1.9", -1},      // FAIL: Negative minor version
		{"5096", -1},        // FAIL: No separators
		{"50.96", -1},       // FAIL: Not enough separators
		{"dummy", -1},       // FAIL: Not numbers
		{"5.0.96.2", -1},    // FAIL: Too many components
		{"5.0.96", 5096},    // OK: 5.0
		{"5.1.72", 5172},    // OK: 5.1
		{"5.5.55", 5555},    // OK: 5.5
		{"ps5.7.20", 5720},  // OK: 5.7 with prefix
		{"5.7.21", 5721},    // OK: 5.7
		{"8.0.0", 8000},     // OK: 8.0
		{"8.0.4", 8004},     // OK: 8.0
		{"8.0.04", 8004},    // OK: 8.0
		{"ma10.2.1", 10201}, // OK: 10.2 with prefix
	}
	//t.Logf("Name: %s\n", t.Name())
	for _, vp := range versions {
		version := vp.version
		expected := vp.port
		port := VersionToPort(version)
		if expected == port {
			t.Logf("ok     %-10s => %5d\n", version, port)
		} else {
			t.Logf("NOT OK %-10s => %5d\n", version, port)
			t.Fail()
		}
	}
}
