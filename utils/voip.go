package utils

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
)

func VoipCall(token string, payload string) error {
	url := "https://api.development.push.apple.com/3/device/" + token

	certPath := "keys/voipCert.pem"
	keyPath := "keys/key.pem"

	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		fmt.Println("Error loading certificate:", err)
		return err
	}

	// Create a custom Transport with TLS configuration
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}
