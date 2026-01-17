package qr

import (
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
)

// QRCode represents a generated QR code
type QRCode struct {
	Data    string
	Size    int
	Content [][]bool // true = black, false = white
}

// QRType defines the type of QR code to generate
type QRType int

const (
	// TypeAccessURL generates a QR code for accessing the code-server
	TypeAccessURL QRType = iota
	// TypeDocumentation generates a QR code linking to documentation
	TypeDocumentation
	// TypeSetupGuide generates a QR code for setup assistance
	TypeSetupGuide
	// TypeConfigExport generates a QR code for sharing configuration
	TypeConfigExport
	// TypeTroubleshooting generates a QR code for troubleshooting help
	TypeTroubleshooting
)

// GeneratorConfig holds configuration for QR code generation
type GeneratorConfig struct {
	// ErrorCorrection level (L=7%, M=15%, Q=25%, H=30%)
	ErrorCorrection string
	// QuietZone is the number of modules for the quiet zone (usually 4)
	QuietZone int
	// InvertColors inverts black/white for terminal display
	InvertColors bool
}

// DefaultConfig returns a default configuration for QR generation
func DefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		ErrorCorrection: "M",
		QuietZone:       2,
		InvertColors:    false,
	}
}

// Generator provides QR code generation functionality
type Generator struct {
	config GeneratorConfig
}

// NewGenerator creates a new QR code generator
func NewGenerator(config GeneratorConfig) *Generator {
	return &Generator{config: config}
}

// NewDefaultGenerator creates a generator with default configuration
func NewDefaultGenerator() *Generator {
	return NewGenerator(DefaultConfig())
}

// GenerateAccessQR generates a QR code for accessing code-server
func (g *Generator) GenerateAccessQR(ip string, port int, https bool) string {
	protocol := "http"
	if https {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s:%d", protocol, ip, port)
	return g.GenerateASCII(url)
}

// GenerateDocQR generates a QR code linking to documentation
func (g *Generator) GenerateDocQR(docPath string) string {
	url := fmt.Sprintf("https://doom-coding.dev/docs/%s", docPath)
	return g.GenerateASCII(url)
}

// GenerateTroubleshootingQR generates a QR code for troubleshooting
func (g *Generator) GenerateTroubleshootingQR(errorCode string) string {
	url := fmt.Sprintf("https://doom-coding.dev/troubleshoot/%s", errorCode)
	return g.GenerateASCII(url)
}

// GenerateExternalServiceQR generates a QR code for external service links
func (g *Generator) GenerateExternalServiceQR(service string) string {
	urls := map[string]string{
		"tailscale-keys": "https://login.tailscale.com/admin/settings/keys",
		"anthropic-keys": "https://console.anthropic.com/account/keys",
		"github-repo":    "https://github.com/doom-coding/doom-coding",
		"termux":         "https://play.google.com/store/apps/details?id=com.termux",
		"blink":          "https://apps.apple.com/app/blink-shell-mosh-ssh-client/id1594898306",
	}

	if url, ok := urls[service]; ok {
		return g.GenerateASCII(url)
	}
	return ""
}

// GenerateASCII generates an ASCII art QR code for terminal display using go-qrcode library
func (g *Generator) GenerateASCII(data string) string {
	// Use go-qrcode library to generate proper QR code
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return fmt.Sprintf("Error generating QR code: %v\nURL: %s", err, data)
	}

	// Get the bitmap
	bitmap := qr.Bitmap()

	// Convert bitmap to ASCII using Unicode block characters
	return g.bitmapToASCII(bitmap)
}

// bitmapToASCII converts a bitmap from go-qrcode library to ASCII art
func (g *Generator) bitmapToASCII(bitmap [][]bool) string {
	var sb strings.Builder

	// Add quiet zone
	quietLine := strings.Repeat("  ", len(bitmap[0])+g.config.QuietZone*2)
	for i := 0; i < g.config.QuietZone; i++ {
		sb.WriteString(quietLine + "\n")
	}

	// Use Unicode block characters for compact display
	// Upper half block: ▀ (U+2580), Lower half block: ▄ (U+2584)
	// Full block: █ (U+2588), Space for white

	// Process two rows at a time for half-height display
	for row := 0; row < len(bitmap); row += 2 {
		// Quiet zone left
		sb.WriteString(strings.Repeat("  ", g.config.QuietZone))

		for col := 0; col < len(bitmap[0]); col++ {
			upper := bitmap[row][col]
			lower := false
			if row+1 < len(bitmap) {
				lower = bitmap[row+1][col]
			}

			if g.config.InvertColors {
				upper = !upper
				lower = !lower
			}

			// Choose character based on upper/lower pattern
			switch {
			case upper && lower:
				sb.WriteString("██")
			case upper && !lower:
				sb.WriteString("▀▀")
			case !upper && lower:
				sb.WriteString("▄▄")
			default:
				sb.WriteString("  ")
			}
		}

		// Quiet zone right
		sb.WriteString(strings.Repeat("  ", g.config.QuietZone))
		sb.WriteString("\n")
	}

	// Add quiet zone bottom
	for i := 0; i < g.config.QuietZone; i++ {
		sb.WriteString(quietLine + "\n")
	}

	return sb.String()
}

// GenerateCompact generates a more compact ASCII QR code
func (g *Generator) GenerateCompact(data string) string {
	// Use go-qrcode library with low error correction for smaller size
	qr, err := qrcode.New(data, qrcode.Low)
	if err != nil {
		return fmt.Sprintf("Error generating compact QR code: %v\nURL: %s", err, data)
	}

	bitmap := qr.Bitmap()
	return g.bitmapToCompactASCII(bitmap)
}

// bitmapToCompactASCII generates a very compact representation
func (g *Generator) bitmapToCompactASCII(bitmap [][]bool) string {
	var sb strings.Builder

	// Use 2x2 block patterns for very compact display
	for row := 0; row < len(bitmap); row += 2 {
		for col := 0; col < len(bitmap[0]); col += 2 {
			topLeft := bitmap[row][col]
			topRight := col+1 < len(bitmap[0]) && bitmap[row][col+1]
			bottomLeft := row+1 < len(bitmap) && bitmap[row+1][col]
			bottomRight := row+1 < len(bitmap) && col+1 < len(bitmap[0]) && bitmap[row+1][col+1]

			// Use quadrant block characters
			char := g.getQuadrantChar(topLeft, topRight, bottomLeft, bottomRight)
			sb.WriteString(char)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// getQuadrantChar returns the appropriate Unicode quadrant character
func (g *Generator) getQuadrantChar(tl, tr, bl, br bool) string {
	if g.config.InvertColors {
		tl, tr, bl, br = !tl, !tr, !bl, !br
	}

	// Unicode quadrant characters
	switch {
	case !tl && !tr && !bl && !br:
		return " "
	case tl && !tr && !bl && !br:
		return "▘"
	case !tl && tr && !bl && !br:
		return "▝"
	case tl && tr && !bl && !br:
		return "▀"
	case !tl && !tr && bl && !br:
		return "▖"
	case tl && !tr && bl && !br:
		return "▌"
	case !tl && tr && bl && !br:
		return "▞"
	case tl && tr && bl && !br:
		return "▛"
	case !tl && !tr && !bl && br:
		return "▗"
	case tl && !tr && !bl && br:
		return "▚"
	case !tl && tr && !bl && br:
		return "▐"
	case tl && tr && !bl && br:
		return "▜"
	case !tl && !tr && bl && br:
		return "▄"
	case tl && !tr && bl && br:
		return "▙"
	case !tl && tr && bl && br:
		return "▟"
	case tl && tr && bl && br:
		return "█"
	default:
		return " "
	}
}

// FormatWithLabel formats a QR code with a descriptive label
func FormatWithLabel(qrCode string, label string) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(qrCode)
	sb.WriteString(fmt.Sprintf("    %s ↑\n", label))
	return sb.String()
}

// GetServiceLabel returns a user-friendly label for a service QR code
func GetServiceLabel(service string) string {
	labels := map[string]string{
		"tailscale-keys": "Scan to create Tailscale auth key",
		"anthropic-keys": "Scan to get Anthropic API key",
		"github-repo":    "Scan to view project on GitHub",
		"termux":         "Scan to install Termux (Android)",
		"blink":          "Scan to install Blink Shell (iOS)",
	}

	if label, ok := labels[service]; ok {
		return label
	}
	return "Scan QR code"
}

// GenerateURLWithFallback generates a QR code with a text URL fallback
func (g *Generator) GenerateURLWithFallback(url string, label string) string {
	var sb strings.Builder

	qr := g.GenerateASCII(url)
	sb.WriteString(qr)
	sb.WriteString(fmt.Sprintf("    %s\n", label))
	sb.WriteString(fmt.Sprintf("    URL: %s\n", url))

	return sb.String()
}
