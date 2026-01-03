package ranking

import (
	"testing"
	"testing/quick"
)

// TestIsUSDTPair tests the IsUSDTPair function with specific examples.
func TestIsUSDTPair(t *testing.T) {
	tests := []struct {
		symbol string
		want   bool
	}{
		{"BTCUSDT", true},
		{"ETHUSDT", true},
		{"SOLUSDT", true},
		{"BTCBUSD", false},
		{"ETHBTC", false},
		{"BNBETH", false},
		{"BTCFDUSD", false},
		{"USDT", false}, // Just "USDT" is not a valid pair
		{"", false},
		{"btcusdt", false}, // Case sensitive
		{"BTCUSDT ", false}, // Trailing space
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			got := IsUSDTPair(tt.symbol)
			if got != tt.want {
				t.Errorf("IsUSDTPair(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

// TestIsUSDTPairProperty tests the property that any string ending with "USDT"
// should return true, and any string not ending with "USDT" should return false.
// Property 1: USDT Pair Filtering
// Validates: Requirements 1.1, 1.2, 2.5
func TestIsUSDTPairProperty(t *testing.T) {
	// Property: For any base currency name, appending "USDT" makes it a USDT pair
	propertyUSDTSuffix := func(base string) bool {
		if len(base) == 0 {
			return true // Skip empty base
		}
		symbol := base + "USDT"
		return IsUSDTPair(symbol)
	}

	if err := quick.Check(propertyUSDTSuffix, nil); err != nil {
		t.Errorf("Property failed: strings ending with USDT should be USDT pairs: %v", err)
	}

	// Property: For any string not ending with "USDT", IsUSDTPair returns false
	propertyNonUSDT := func(s string) bool {
		// Skip strings that actually end with USDT
		if len(s) >= 4 && s[len(s)-4:] == "USDT" {
			return true
		}
		return !IsUSDTPair(s)
	}

	if err := quick.Check(propertyNonUSDT, nil); err != nil {
		t.Errorf("Property failed: strings not ending with USDT should not be USDT pairs: %v", err)
	}
}

// TestNonUSDTPairsExcluded tests that common non-USDT quote currencies are excluded.
func TestNonUSDTPairsExcluded(t *testing.T) {
	nonUSDTPairs := []string{
		"ETHBTC",   // BTC quoted
		"BNBBTC",   // BTC quoted
		"SOLETH",   // ETH quoted
		"ADABNB",   // BNB quoted
		"BTCFDUSD", // FDUSD quoted
		"ETHFDUSD", // FDUSD quoted
		"BTCBUSD",  // BUSD quoted
		"BTCTUSD",  // TUSD quoted
	}

	for _, pair := range nonUSDTPairs {
		if IsUSDTPair(pair) {
			t.Errorf("IsUSDTPair(%q) = true, want false (non-USDT pair should be excluded)", pair)
		}
	}
}
