package styles

import "github.com/charmbracelet/lipgloss"

// Brand colors for Doom Coding
var (
	ForestGreen  = lipgloss.Color("#2E521D")
	TanBrown     = lipgloss.Color("#7C5E46")
	LightGreen   = lipgloss.Color("#4A7C34")
	DarkGreen    = lipgloss.Color("#1A3010")
	Cream        = lipgloss.Color("#F5F5DC")
	White        = lipgloss.Color("#FFFFFF")
	Gray         = lipgloss.Color("#888888")
	DarkGray     = lipgloss.Color("#666666")
	Red          = lipgloss.Color("#FF6B6B")
	Green        = lipgloss.Color("#69DB7C")
	Yellow       = lipgloss.Color("#FFD93D")
	Blue         = lipgloss.Color("#4DABF7")
)

// Text styles
var (
	Title = lipgloss.NewStyle().
		Foreground(ForestGreen).
		Bold(true).
		MarginBottom(1)

	Subtitle = lipgloss.NewStyle().
		Foreground(TanBrown).
		MarginBottom(1)

	Heading = lipgloss.NewStyle().
		Foreground(LightGreen).
		Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(White)

	Dimmed = lipgloss.NewStyle().
		Foreground(Gray)

	Disabled = lipgloss.NewStyle().
		Foreground(DarkGray)

	Selected = lipgloss.NewStyle().
		Foreground(LightGreen).
		Bold(true)

	Focused = lipgloss.NewStyle().
		Foreground(ForestGreen).
		Bold(true)

	Success = lipgloss.NewStyle().
		Foreground(Green)

	Error = lipgloss.NewStyle().
		Foreground(Red)

	Warning = lipgloss.NewStyle().
		Foreground(Yellow)

	Info = lipgloss.NewStyle().
		Foreground(Blue)
)

// Container styles
var (
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ForestGreen).
		Padding(1, 2)

	FocusedBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(LightGreen).
		Padding(1, 2)

	Card = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(TanBrown).
		Padding(0, 1)

	StatusBar = lipgloss.NewStyle().
		Background(ForestGreen).
		Foreground(White).
		Padding(0, 1)

	HelpBar = lipgloss.NewStyle().
		Foreground(Gray).
		MarginTop(1)
)

// Input styles
var (
	InputPrompt = lipgloss.NewStyle().
		Foreground(TanBrown)

	InputText = lipgloss.NewStyle().
		Foreground(White)

	InputPlaceholder = lipgloss.NewStyle().
		Foreground(DarkGray)

	InputFocused = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(LightGreen).
		Padding(0, 1)

	InputBlurred = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(DarkGray).
		Padding(0, 1)
)

// List styles
var (
	ListItem = lipgloss.NewStyle().
		PaddingLeft(2)

	ListItemSelected = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(LightGreen).
		Bold(true)

	ListItemDisabled = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(DarkGray)

	Cursor = lipgloss.NewStyle().
		Foreground(LightGreen).
		Bold(true)

	CheckboxChecked = lipgloss.NewStyle().
		Foreground(Green).
		SetString("[✓]")

	CheckboxUnchecked = lipgloss.NewStyle().
		Foreground(Gray).
		SetString("[ ]")

	CheckboxDisabled = lipgloss.NewStyle().
		Foreground(DarkGray).
		SetString("[-]")

	RadioSelected = lipgloss.NewStyle().
		Foreground(Green).
		SetString("(●)")

	RadioUnselected = lipgloss.NewStyle().
		Foreground(Gray).
		SetString("( )")
)

// Progress styles
var (
	ProgressBar = lipgloss.NewStyle().
		Foreground(LightGreen)

	ProgressBarBackground = lipgloss.NewStyle().
		Foreground(DarkGray)

	ProgressLabel = lipgloss.NewStyle().
		Foreground(TanBrown)

	ProgressPercent = lipgloss.NewStyle().
		Foreground(LightGreen).
		Bold(true)
)

// Badge styles
var (
	BadgeSuccess = lipgloss.NewStyle().
		Background(Green).
		Foreground(DarkGreen).
		Padding(0, 1).
		Bold(true)

	BadgeError = lipgloss.NewStyle().
		Background(Red).
		Foreground(White).
		Padding(0, 1).
		Bold(true)

	BadgeWarning = lipgloss.NewStyle().
		Background(Yellow).
		Foreground(DarkGreen).
		Padding(0, 1).
		Bold(true)

	BadgeInfo = lipgloss.NewStyle().
		Background(Blue).
		Foreground(White).
		Padding(0, 1).
		Bold(true)
)

// Banner renders the Doom Coding ASCII art banner
func Banner() string {
	banner := `
    ██████╗  ██████╗  ██████╗ ███╗   ███╗
    ██╔══██╗██╔═══██╗██╔═══██╗████╗ ████║
    ██║  ██║██║   ██║██║   ██║██╔████╔██║
    ██║  ██║██║   ██║██║   ██║██║╚██╔╝██║
    ██████╔╝╚██████╔╝╚██████╔╝██║ ╚═╝ ██║
    ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝     ╚═╝
     ██████╗ ██████╗ ██████╗ ██╗███╗   ██╗ ██████╗
    ██╔════╝██╔═══██╗██╔══██╗██║████╗  ██║██╔════╝
    ██║     ██║   ██║██║  ██║██║██╔██╗ ██║██║  ███╗
    ██║     ██║   ██║██║  ██║██║██║╚██╗██║██║   ██║
    ╚██████╗╚██████╔╝██████╔╝██║██║ ╚████║╚██████╔╝
     ╚═════╝ ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝ ╚═════╝
`
	return lipgloss.NewStyle().Foreground(ForestGreen).Bold(true).Render(banner)
}

// Icons
const (
	IconCheck    = "✓"
	IconCross    = "✗"
	IconWarning  = "⚠"
	IconInfo     = "ℹ"
	IconArrow    = "▸"
	IconDot      = "●"
	IconCircle   = "○"
	IconDocker   = "🐳"
	IconNetwork  = "🌐"
	IconHome     = "🏠"
	IconBolt     = "⚡"
	IconLock     = "🔒"
	IconKey      = "🔑"
	IconTerminal = "💻"
	IconFolder   = "📁"
	IconGear     = "⚙"
	IconRocket   = "🚀"
)

// StatusIcon returns the appropriate icon for a status
func StatusIcon(success bool) string {
	if success {
		return Success.Render(IconCheck)
	}
	return Error.Render(IconCross)
}

// Center centers content within a given width
func Center(content string, width int) string {
	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, content)
}

// CenterVertical centers content vertically
func CenterVertical(content string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
