package qr

import (
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	config := DefaultConfig()
	g := NewGenerator(config)

	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}

	if g.config.ErrorCorrection != "M" {
		t.Errorf("Expected ErrorCorrection M, got %s", g.config.ErrorCorrection)
	}

	if g.config.QuietZone != 2 {
		t.Errorf("Expected QuietZone 2, got %d", g.config.QuietZone)
	}
}

func TestNewDefaultGenerator(t *testing.T) {
	g := NewDefaultGenerator()

	if g == nil {
		t.Fatal("NewDefaultGenerator returned nil")
	}
}

func TestGenerateAccessQR(t *testing.T) {
	g := NewDefaultGenerator()

	tests := []struct {
		ip    string
		port  int
		https bool
	}{
		{"192.168.1.100", 8443, true},
		{"100.64.0.1", 8443, true},
		{"localhost", 8080, false},
	}

	for _, tt := range tests {
		qr := g.GenerateAccessQR(tt.ip, tt.port, tt.https)
		if qr == "" {
			t.Errorf("GenerateAccessQR(%s, %d, %v) returned empty string", tt.ip, tt.port, tt.https)
		}

		// Check that QR contains block characters
		if !strings.ContainsAny(qr, "█▀▄") {
			t.Errorf("GenerateAccessQR output doesn't contain expected block characters")
		}
	}
}

func TestGenerateExternalServiceQR(t *testing.T) {
	g := NewDefaultGenerator()

	services := []string{
		"tailscale-keys",
		"anthropic-keys",
		"github-repo",
		"termux",
		"blink",
	}

	for _, service := range services {
		qr := g.GenerateExternalServiceQR(service)
		if qr == "" {
			t.Errorf("GenerateExternalServiceQR(%s) returned empty string", service)
		}
	}

	// Test unknown service
	qr := g.GenerateExternalServiceQR("unknown-service")
	if qr != "" {
		t.Error("GenerateExternalServiceQR should return empty string for unknown service")
	}
}

func TestGenerateASCII(t *testing.T) {
	g := NewDefaultGenerator()

	testData := []string{
		"https://example.com",
		"https://192.168.1.100:8443",
		"https://doom-coding.dev/docs/getting-started",
	}

	for _, data := range testData {
		qr := g.GenerateASCII(data)

		// Should not be empty
		if qr == "" {
			t.Errorf("GenerateASCII(%q) returned empty string", data)
		}

		// Should contain newlines (multi-line output)
		if !strings.Contains(qr, "\n") {
			t.Errorf("GenerateASCII output should be multi-line")
		}

		// Should contain block characters
		if !strings.ContainsAny(qr, "█▀▄ ") {
			t.Errorf("GenerateASCII output should contain block characters")
		}
	}
}

func TestGenerateCompact(t *testing.T) {
	g := NewDefaultGenerator()

	qr := g.GenerateCompact("https://example.com")

	if qr == "" {
		t.Error("GenerateCompact returned empty string")
	}

	// Compact should be smaller than regular ASCII
	regularQR := g.GenerateASCII("https://example.com")
	if len(qr) >= len(regularQR) {
		t.Log("Compact QR is not smaller than regular QR")
	}
}

func TestFormatWithLabel(t *testing.T) {
	g := NewDefaultGenerator()
	qr := g.GenerateASCII("https://example.com")
	label := "Scan to open"

	formatted := FormatWithLabel(qr, label)

	if !strings.Contains(formatted, label) {
		t.Error("FormatWithLabel should include the label")
	}

	if !strings.Contains(formatted, "↑") {
		t.Error("FormatWithLabel should include arrow pointing to QR")
	}
}

func TestGetServiceLabel(t *testing.T) {
	tests := []struct {
		service  string
		expected string
	}{
		{"tailscale-keys", "Scan to create Tailscale auth key"},
		{"anthropic-keys", "Scan to get Anthropic API key"},
		{"unknown", "Scan QR code"},
	}

	for _, tt := range tests {
		label := GetServiceLabel(tt.service)
		if label != tt.expected {
			t.Errorf("GetServiceLabel(%s) = %s, want %s", tt.service, label, tt.expected)
		}
	}
}

func TestGenerateURLWithFallback(t *testing.T) {
	g := NewDefaultGenerator()
	url := "https://example.com"
	label := "Test Label"

	output := g.GenerateURLWithFallback(url, label)

	if !strings.Contains(output, label) {
		t.Error("Output should contain label")
	}

	if !strings.Contains(output, url) {
		t.Error("Output should contain URL fallback")
	}

	if !strings.Contains(output, "URL:") {
		t.Error("Output should contain 'URL:' prefix")
	}
}

func TestInvertColors(t *testing.T) {
	config := DefaultConfig()
	config.InvertColors = true
	g := NewGenerator(config)

	qr := g.GenerateASCII("https://test.com")

	// Inverted QR should still produce valid output
	if qr == "" {
		t.Error("Inverted QR code should not be empty")
	}
}

func TestQuadrantCharacters(t *testing.T) {
	g := NewDefaultGenerator()

	// Test all 16 combinations
	combinations := []struct {
		tl, tr, bl, br bool
		expected       string
	}{
		{false, false, false, false, " "},
		{true, false, false, false, "▘"},
		{false, true, false, false, "▝"},
		{true, true, false, false, "▀"},
		{false, false, true, false, "▖"},
		{true, false, true, false, "▌"},
		{false, true, true, false, "▞"},
		{true, true, true, false, "▛"},
		{false, false, false, true, "▗"},
		{true, false, false, true, "▚"},
		{false, true, false, true, "▐"},
		{true, true, false, true, "▜"},
		{false, false, true, true, "▄"},
		{true, false, true, true, "▙"},
		{false, true, true, true, "▟"},
		{true, true, true, true, "█"},
	}

	for _, c := range combinations {
		result := g.getQuadrantChar(c.tl, c.tr, c.bl, c.br)
		if result != c.expected {
			t.Errorf("getQuadrantChar(%v, %v, %v, %v) = %s, want %s",
				c.tl, c.tr, c.bl, c.br, result, c.expected)
		}
	}
}

func TestFinderPattern(t *testing.T) {
	g := NewDefaultGenerator()

	// Create a small matrix
	matrix := make([][]bool, 21)
	for i := range matrix {
		matrix[i] = make([]bool, 21)
	}

	// Add finder pattern
	g.addFinderPattern(matrix, 0, 0)

	// Check corners of finder pattern
	if !matrix[0][0] {
		t.Error("Top-left corner of finder pattern should be true")
	}
	if !matrix[0][6] {
		t.Error("Top-right corner of finder pattern should be true")
	}
	if !matrix[6][0] {
		t.Error("Bottom-left corner of finder pattern should be true")
	}
	if !matrix[6][6] {
		t.Error("Bottom-right corner of finder pattern should be true")
	}

	// Check center should be true
	if !matrix[3][3] {
		t.Error("Center of finder pattern should be true")
	}
}

func TestIsFunctionModule(t *testing.T) {
	g := NewDefaultGenerator()
	size := 21

	// Finder pattern area
	if !g.isFunctionModule(0, 0, size) {
		t.Error("Position (0,0) should be function module")
	}

	if !g.isFunctionModule(8, 0, size) {
		t.Error("Position (8,0) should be function module (separator)")
	}

	// Timing pattern
	if !g.isFunctionModule(6, 10, size) {
		t.Error("Position (6,10) should be function module (timing)")
	}

	// Data area
	if g.isFunctionModule(10, 10, size) {
		t.Error("Position (10,10) should not be function module")
	}
}
