package certificate

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CertificateCodeGenerator handles generation of unique certificate codes
type CertificateCodeGenerator struct{}

// NewCertificateCodeGenerator creates a new certificate code generator
func NewCertificateCodeGenerator() *CertificateCodeGenerator {
	return &CertificateCodeGenerator{}
}

// GenerateCode generates a unique certificate code
// Format: CERT-{YYYYMMDD}-{RANDOM_HEX}
func (g *CertificateCodeGenerator) GenerateCode() (string, error) {
	// Get current date in YYYYMMDD format
	dateStr := time.Now().Format("20060102")
	
	// Generate 8 random bytes and convert to hex
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomHex := hex.EncodeToString(randomBytes)
	
	// Format: CERT-YYYYMMDD-XXXXXXXX
	code := fmt.Sprintf("CERT-%s-%s", dateStr, strings.ToUpper(randomHex))
	
	return code, nil
}

// GenerateCodeWithUserID generates a unique certificate code with user ID prefix
// Format: CERT-{USER_ID_SHORT}-{YYYYMMDD}-{RANDOM_HEX}
func (g *CertificateCodeGenerator) GenerateCodeWithUserID(userID uuid.UUID) (string, error) {
	// Get first 8 characters of user ID
	userIDStr := strings.ReplaceAll(userID.String(), "-", "")[:8]
	
	// Get current date in YYYYMMDD format
	dateStr := time.Now().Format("20060102")
	
	// Generate 6 random bytes and convert to hex
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomHex := hex.EncodeToString(randomBytes)
	
	// Format: CERT-{USER_ID_SHORT}-{YYYYMMDD}-{RANDOM_HEX}
	code := fmt.Sprintf("CERT-%s-%s-%s", strings.ToUpper(userIDStr), dateStr, strings.ToUpper(randomHex))
	
	return code, nil
}

// ValidateCodeFormat validates if a certificate code has the correct format
func (g *CertificateCodeGenerator) ValidateCodeFormat(code string) bool {
	// Check if code starts with CERT- and has proper format
	if !strings.HasPrefix(code, "CERT-") {
		return false
	}
	
	// Split by dashes and check parts
	parts := strings.Split(code, "-")
	if len(parts) < 3 {
		return false
	}
	
	// Check if date part is valid (8 digits)
	if len(parts[1]) != 8 {
		return false
	}
	
	// Check if random part exists and is hex
	if len(parts[2]) < 8 {
		return false
	}
	
	return true
}

// ExtractDateFromCode extracts the date from a certificate code
func (g *CertificateCodeGenerator) ExtractDateFromCode(code string) (time.Time, error) {
	if !g.ValidateCodeFormat(code) {
		return time.Time{}, fmt.Errorf("invalid certificate code format")
	}
	
	parts := strings.Split(code, "-")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid certificate code format")
	}
	
	dateStr := parts[1]
	return time.Parse("20060102", dateStr)
}
