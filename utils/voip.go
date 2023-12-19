package utils

import (
	"crypto/tls"
	"fmt"
	"io"
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

	fmt.Println("Status:", resp.Status)
	fmt.Println("StatusCode:", resp.StatusCode)
	fmt.Println("Headers:", resp.Header)

	// Для тела ответа также нужно использовать io.Reader из resp.Body
	// Например, считать его в []byte и вывести
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	} else {
		fmt.Println("Body:", string(bodyBytes))
	}

	defer resp.Body.Close()

	return nil
}
