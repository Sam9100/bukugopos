package wasender

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
)

// SendMessage sends a WhatsApp message via configured provider (Fonnte)
func SendMessage(to, message string) error {
	switch config.WAProvider {
	case "fonnte":
		return sendViaFonnte(to, message)
	default:
		return fmt.Errorf("unknown WA provider: %s", config.WAProvider)
	}
}

// sendViaFonnte sends message using Fonnte API
// API Docs: https://docs.fonnte.com/mengirim-pesan-api/
// NOTE: Fonnte uses form-urlencoded, NOT JSON!
func sendViaFonnte(to, message string) error {
	token := config.GetFonnteToken()
	if token == "" {
		return fmt.Errorf("FONNTETOKEN not set")
	}

	// For groups, log for debugging
	isGroup := strings.Contains(to, "@g.us")
	if isGroup {
		fmt.Printf("📢 Sending to GROUP: %s\n", to)
	}

	// Prepare form data (NOT JSON!)
	formData := url.Values{}
	formData.Set("target", to)
	formData.Set("message", message)
	formData.Set("countryCode", "62")

	fmt.Printf("📤 Fonnte sending to: %s, formData: %s\n", to, formData.Encode())

	// Create HTTP request with form-urlencoded body
	req, err := http.NewRequest("POST", config.FonnteAPIURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log raw response for debugging
	fmt.Printf("📥 Fonnte response (status=%d): %s\n", resp.StatusCode, string(body))

	// Parse response
	var fonnteResp model.FonnteResponse
	if err := json.Unmarshal(body, &fonnteResp); err != nil {
		fmt.Printf("⚠️ Fonnte response parse error: %v, raw: %s\n", err, string(body))
		return nil // Don't fail if response can't be parsed
	}

	if !fonnteResp.Status {
		return fmt.Errorf("fonnte error: %s (reason: %s)", fonnteResp.Detail, fonnteResp.Reason)
	}

	fmt.Printf("✅ Fonnte sent to %s: %s\n", to, message[:min(50, len(message))])
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
