package config

import "testing"

func TestQRCredentialsFallsBackToMainPair(t *testing.T) {
	t.Parallel()

	// The common case: CU NEX issued one credential pair, which serves both
	// /profile and qrcodeinfo_for_all. Only two env vars need to be set.
	main := LLEConfig{ClientID: "id", ClientSecret: "secret"}
	if gotID, gotSecret := main.QRCredentials(); gotID != "id" || gotSecret != "secret" {
		t.Fatalf("QRCredentials() = (%q, %q), want the main pair", gotID, gotSecret)
	}

	// If CU NEX later issues a QR-specific pair, setting both halves takes over.
	split := LLEConfig{
		ClientID: "id", ClientSecret: "secret",
		QRClientID: "qr-id", QRClientSecret: "qr-secret",
	}
	if gotID, gotSecret := split.QRCredentials(); gotID != "qr-id" || gotSecret != "qr-secret" {
		t.Fatalf("QRCredentials() = (%q, %q), want the QR pair", gotID, gotSecret)
	}

	// A half-configured override is a misconfiguration, not an intent to use a
	// separate credential — fall back rather than send an empty secret.
	for _, partial := range []LLEConfig{
		{ClientID: "id", ClientSecret: "secret", QRClientID: "qr-id"},
		{ClientID: "id", ClientSecret: "secret", QRClientSecret: "qr-secret"},
		{ClientID: "id", ClientSecret: "secret", QRClientID: "  ", QRClientSecret: "  "},
	} {
		if gotID, gotSecret := partial.QRCredentials(); gotID != "id" || gotSecret != "secret" {
			t.Errorf("QRCredentials() with a half-set override = (%q, %q), want the main pair", gotID, gotSecret)
		}
	}
}
