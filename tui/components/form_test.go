package components

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewForm(t *testing.T) {
	fields := []FormField{
		{Label: "Username", Placeholder: "Enter username", Required: true},
		{Label: "Password", Placeholder: "Enter password", Secret: true},
		{Label: "Email", Placeholder: "user@example.com", Help: "Your email address"},
	}

	form := NewForm("Registration Form", fields)

	if form.Title != "Registration Form" {
		t.Errorf("Expected title 'Registration Form', got %q", form.Title)
	}

	if len(form.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(form.Fields))
	}

	if form.FocusIndex != 0 {
		t.Errorf("Expected focus index at 0, got %d", form.FocusIndex)
	}

	if !form.Focused {
		t.Error("Expected Focused to be true by default")
	}

	if form.Submitted {
		t.Error("Expected Submitted to be false initially")
	}

	if form.Errors == nil {
		t.Error("Errors map should be initialized")
	}
}

func TestFormInit(t *testing.T) {
	form := NewForm("Test", []FormField{
		{Label: "Field 1"},
	})

	cmd := form.Init()
	// Init returns textinput.Blink for cursor blinking
	if cmd == nil {
		t.Error("Init should return a command for text input blinking")
	}
}

func TestFormFieldNavigation(t *testing.T) {
	fields := []FormField{
		{Label: "Field 1"},
		{Label: "Field 2"},
		{Label: "Field 3"},
	}
	form := NewForm("Test", fields)

	// Test tab navigation
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	form, _ = form.Update(tabMsg)
	if form.FocusIndex != 1 {
		t.Errorf("Tab should move to field 1, got %d", form.FocusIndex)
	}

	form, _ = form.Update(tabMsg)
	if form.FocusIndex != 2 {
		t.Errorf("Tab should move to field 2, got %d", form.FocusIndex)
	}

	form, _ = form.Update(tabMsg)
	if form.FocusIndex != 0 {
		t.Errorf("Tab should wrap to field 0, got %d", form.FocusIndex)
	}

	// Test shift+tab navigation (backward)
	shiftTabMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	form, _ = form.Update(shiftTabMsg)
	if form.FocusIndex != 2 {
		t.Errorf("Shift+Tab should move to field 2, got %d", form.FocusIndex)
	}

	// Test down key
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	form, _ = form.Update(downMsg)
	if form.FocusIndex != 0 {
		t.Errorf("Down should wrap to field 0, got %d", form.FocusIndex)
	}

	// Test up key
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	form, _ = form.Update(upMsg)
	if form.FocusIndex != 2 {
		t.Errorf("Up should move to field 2, got %d", form.FocusIndex)
	}
}

func TestFormValidateRequired(t *testing.T) {
	fields := []FormField{
		{Label: "Required Field", Required: true},
		{Label: "Optional Field", Required: false},
	}
	form := NewForm("Test", fields)

	// Should fail validation with empty required field
	valid := form.Validate()
	if valid {
		t.Error("Validation should fail with empty required field")
	}

	if _, exists := form.Errors[0]; !exists {
		t.Error("Error should be set for required field")
	}

	// Set value and revalidate
	form.SetValue(0, "some value")
	valid = form.Validate()
	if !valid {
		t.Error("Validation should pass with required field filled")
	}
}

func TestFormValidateCustomValidator(t *testing.T) {
	customValidator := func(value string) error {
		if len(value) < 5 {
			return errors.New("must be at least 5 characters")
		}
		return nil
	}

	fields := []FormField{
		{Label: "Custom", Validator: customValidator},
	}
	form := NewForm("Test", fields)

	// Set short value
	form.SetValue(0, "abc")
	valid := form.Validate()
	if valid {
		t.Error("Validation should fail with short value")
	}

	if errMsg, exists := form.Errors[0]; !exists || !strings.Contains(errMsg, "at least 5") {
		t.Errorf("Expected specific error message, got %q", errMsg)
	}

	// Set valid value
	form.SetValue(0, "abcdef")
	valid = form.Validate()
	if !valid {
		t.Error("Validation should pass with valid value")
	}
}

func TestFormGetValues(t *testing.T) {
	fields := []FormField{
		{Label: "Username", Value: "testuser"},
		{Label: "Email", Value: "test@example.com"},
	}
	form := NewForm("Test", fields)

	// Note: Values are set via the input model, not directly
	// For testing, we need to use SetValue
	form.SetValue(0, "testuser")
	form.SetValue(1, "test@example.com")

	values := form.GetValues()

	if values["Username"] != "testuser" {
		t.Errorf("Expected Username='testuser', got %q", values["Username"])
	}
	if values["Email"] != "test@example.com" {
		t.Errorf("Expected Email='test@example.com', got %q", values["Email"])
	}
}

func TestFormGetValue(t *testing.T) {
	fields := []FormField{
		{Label: "Username"},
		{Label: "Email"},
	}
	form := NewForm("Test", fields)

	form.SetValue(0, "testuser")

	if form.GetValue("Username") != "testuser" {
		t.Errorf("Expected 'testuser', got %q", form.GetValue("Username"))
	}

	// Non-existent field
	if form.GetValue("NonExistent") != "" {
		t.Errorf("Expected empty string for non-existent field, got %q", form.GetValue("NonExistent"))
	}
}

func TestFormSetValue(t *testing.T) {
	fields := []FormField{
		{Label: "Field 1"},
	}
	form := NewForm("Test", fields)

	form.SetValue(0, "test value")
	if form.Fields[0].Value != "test value" {
		t.Errorf("Expected 'test value', got %q", form.Fields[0].Value)
	}

	// Test with invalid indices
	form.SetValue(-1, "invalid") // Should not panic
	form.SetValue(100, "invalid") // Should not panic
}

func TestFormFocusAndBlur(t *testing.T) {
	fields := []FormField{
		{Label: "Field 1"},
	}
	form := NewForm("Test", fields)

	// Initially focused
	if !form.Focused {
		t.Error("Form should be focused initially")
	}

	// Blur
	form.Blur()
	if form.Focused {
		t.Error("Form should not be focused after Blur")
	}

	// Focus
	form.Focus()
	if !form.Focused {
		t.Error("Form should be focused after Focus")
	}
}

func TestFormReset(t *testing.T) {
	fields := []FormField{
		{Label: "Field 1"},
		{Label: "Field 2"},
	}
	form := NewForm("Test", fields)

	form.SetValue(0, "some value")
	form.FocusIndex = 1
	form.Submitted = true
	form.Errors[0] = "some error"

	form.Reset()

	if form.FocusIndex != 0 {
		t.Errorf("FocusIndex should be 0 after reset, got %d", form.FocusIndex)
	}
	if form.Submitted {
		t.Error("Submitted should be false after reset")
	}
	if len(form.Errors) != 0 {
		t.Error("Errors should be empty after reset")
	}
}

func TestFormUpdateNotFocused(t *testing.T) {
	fields := []FormField{
		{Label: "Field 1"},
	}
	form := NewForm("Test", fields)
	form.Focused = false

	// Messages should not be processed when not focused
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	form, _ = form.Update(tabMsg)
	if form.FocusIndex != 0 {
		t.Error("Focus index should not change when not focused")
	}
}

func TestFormView(t *testing.T) {
	fields := []FormField{
		{Label: "Username", Placeholder: "Enter username", Required: true, Help: "Your username"},
		{Label: "Password", Placeholder: "Enter password", Secret: true},
	}
	form := NewForm("Login Form", fields)

	view := form.View()

	// Check title is present
	if !strings.Contains(view, "Login Form") {
		t.Error("View should contain title")
	}

	// Check labels are present
	if !strings.Contains(view, "Username") {
		t.Error("View should contain Username label")
	}
	if !strings.Contains(view, "Password") {
		t.Error("View should contain Password label")
	}

	// Check help text is present
	if !strings.Contains(view, "Your username") {
		t.Error("View should contain help text")
	}
}

func TestFormViewWithErrors(t *testing.T) {
	fields := []FormField{
		{Label: "Username", Required: true},
	}
	form := NewForm("Test", fields)

	// Trigger validation error
	form.Validate()

	view := form.View()

	// Error should be displayed
	if !strings.Contains(view, "required") {
		t.Error("View should show validation error")
	}
}

func TestFormViewNoTitle(t *testing.T) {
	fields := []FormField{
		{Label: "Field 1"},
	}
	form := NewForm("", fields)

	view := form.View()

	// Should still render without title
	if !strings.Contains(view, "Field 1") {
		t.Error("View should contain Field 1")
	}
}

func TestFormEnterSubmit(t *testing.T) {
	fields := []FormField{
		{Label: "Field 1"},
		{Label: "Field 2"},
	}
	form := NewForm("Test", fields)

	// Set values to pass validation
	form.SetValue(0, "value1")
	form.SetValue(1, "value2")

	// Move to last field
	form.FocusIndex = 1

	// Press enter on last field
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	form, _ = form.Update(enterMsg)

	if !form.Submitted {
		t.Error("Form should be submitted after enter on last field with valid values")
	}
}

func TestFormEnterSubmitInvalid(t *testing.T) {
	fields := []FormField{
		{Label: "Field 1", Required: true},
	}
	form := NewForm("Test", fields)

	// Don't set value - validation should fail
	form.FocusIndex = 0

	// Press enter on last field (also first in this case)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	form, _ = form.Update(enterMsg)

	if form.Submitted {
		t.Error("Form should not be submitted with invalid values")
	}
}

func TestValidateNotEmpty(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"hello", false},
		{"  hello  ", false},
		{"", true},
		{"   ", true},
		{"\t\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := ValidateNotEmpty(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotEmpty(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	validator := ValidateMinLength(5)

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"12345", false},
		{"123456", false},
		{"1234", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := validator(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMinLength(5)(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTailscaleKey(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"tskey-auth-123456", false},
		{"tskey-client-abcdef", false},
		{"invalid-key", true},
		{"", true},
		{"ts-key-missing", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := ValidateTailscaleKey(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTailscaleKey(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAnthropicKey(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"sk-ant-api03-abc123", false},
		{"sk-ant-test", false},
		{"invalid-key", true},
		{"", true},
		{"sk-openai-key", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := ValidateAnthropicKey(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAnthropicKey(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestFormEmptyFields(t *testing.T) {
	form := NewForm("Empty Form", []FormField{})

	// Should handle empty fields gracefully
	form.Validate() // Should not panic
	form.Focus()    // Should not panic
	form.Blur()     // Should not panic
	form.Reset()    // Should not panic

	view := form.View()
	if !strings.Contains(view, "Empty Form") {
		t.Error("View should contain title even with no fields")
	}
}

func TestFormSecretField(t *testing.T) {
	fields := []FormField{
		{Label: "Password", Secret: true, Value: "secretpass"},
	}
	form := NewForm("Test", fields)

	// The secret field should have EchoMode set (tested via initialization)
	// The actual echo mode is handled by textinput internally
	view := form.View()

	// Password field should be present
	if !strings.Contains(view, "Password") {
		t.Error("View should contain Password label")
	}
}

func TestFormFieldWithInitialValue(t *testing.T) {
	fields := []FormField{
		{Label: "Username", Value: "initialuser"},
	}
	form := NewForm("Test", fields)

	// The initial value should be set in the input
	value := form.GetValue("Username")
	if value != "initialuser" {
		t.Errorf("Expected initial value 'initialuser', got %q", value)
	}
}

func TestFormValidatorWithEmptyValue(t *testing.T) {
	// Custom validator that should not be called for empty non-required fields
	callCount := 0
	customValidator := func(value string) error {
		callCount++
		return nil
	}

	fields := []FormField{
		{Label: "Optional", Validator: customValidator, Required: false},
	}
	form := NewForm("Test", fields)

	// With empty value, validator should not be called (see form.go line 217)
	form.Validate()

	if callCount > 0 {
		t.Error("Validator should not be called for empty non-required field")
	}
}
