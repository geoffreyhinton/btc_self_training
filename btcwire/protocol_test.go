package btcwire_test

import (
	"github.com/geoffreyhinton/btc_self_training/btcwire"
	"testing"
)

// TestServiceFlagStringer tests the stringized output for service flag types.
func TestServiceFlagStringer(t *testing.T) {
	tests := []struct {
		in   btcwire.ServiceFlag
		want string
	}{
		{0, "0x0"},
		{btcwire.SFNodeNetwork, "SFNodeNetwork"},
		{0xffffffff, "SFNodeNetwork|0xfffffffe"},
		{btcwire.ServiceFlag(0x2), "0x2"},
		{btcwire.ServiceFlag(0x3), "SFNodeNetwork|0x2"},
		{btcwire.ServiceFlag(0x8000000000000000), "0x8000000000000000"},
	}
	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}
