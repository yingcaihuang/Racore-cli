package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "SSL certificate management",
	Example: `  # List all certificates
  racore-cli cert list

  # Upload a certificate
  racore-cli cert upload --name "My Cert" --cert-file ./cert.pem --key-file ./key.pem

  # Update a certificate
  racore-cli cert update --id 123 --cert-file ./new-cert.pem --key-file ./new-key.pem

  # Apply for AWS certificate
  racore-cli cert apply-aws --domain example.com

  # Get validation info
  racore-cli cert validation-info --id 123`,
}

var certListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all SSL certificates",
	Example: `  racore-cli cert list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := executeCertList()
		if err != nil {
			return err
		}
		fmt.Print(result)
		return nil
	},
}

var certUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a new SSL certificate",
	Example: `  racore-cli cert upload --name "My Cert" --cert-file ./cert.pem --key-file ./key.pem`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		certFile, _ := cmd.Flags().GetString("cert-file")
		keyFile, _ := cmd.Flags().GetString("key-file")

		result, err := executeCertUpload(name, certFile, keyFile)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var certUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing SSL certificate",
	Example: `  racore-cli cert update --id 123 --cert-file ./new-cert.pem --key-file ./new-key.pem`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		certFile, _ := cmd.Flags().GetString("cert-file")
		keyFile, _ := cmd.Flags().GetString("key-file")

		result, err := executeCertUpdate(id, certFile, keyFile)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var certApplyAWSCmd = &cobra.Command{
	Use:     "apply-aws",
	Short:   "Apply for an AWS certificate",
	Example: `  racore-cli cert apply-aws --domain example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")

		result, err := executeCertApplyAWS(domain)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var certValidationInfoCmd = &cobra.Command{
	Use:     "validation-info",
	Short:   "Get certificate validation information",
	Example: `  racore-cli cert validation-info --id 123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")

		result, err := executeCertValidationInfo(id)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func init() {
	// Register flags for cert upload
	certUploadCmd.Flags().String("name", "", "Display name for the certificate")
	certUploadCmd.Flags().String("cert-file", "", "Path to PEM-encoded certificate file")
	certUploadCmd.Flags().String("key-file", "", "Path to PEM-encoded private key file")
	_ = certUploadCmd.MarkFlagRequired("name")
	_ = certUploadCmd.MarkFlagRequired("cert-file")
	_ = certUploadCmd.MarkFlagRequired("key-file")

	// Register flags for cert update
	certUpdateCmd.Flags().String("id", "", "Certificate ID to update")
	certUpdateCmd.Flags().String("cert-file", "", "Path to new PEM-encoded certificate file")
	certUpdateCmd.Flags().String("key-file", "", "Path to new PEM-encoded private key file")
	_ = certUpdateCmd.MarkFlagRequired("id")
	_ = certUpdateCmd.MarkFlagRequired("cert-file")
	_ = certUpdateCmd.MarkFlagRequired("key-file")

	// Register flags for cert apply-aws
	certApplyAWSCmd.Flags().String("domain", "", "Domain to apply AWS certificate for (e.g., example.com)")
	_ = certApplyAWSCmd.MarkFlagRequired("domain")

	// Register flags for cert validation-info
	certValidationInfoCmd.Flags().String("id", "", "Certificate ID to get validation info for")
	_ = certValidationInfoCmd.MarkFlagRequired("id")

	// Add subcommands to cert
	certCmd.AddCommand(certListCmd)
	certCmd.AddCommand(certUploadCmd)
	certCmd.AddCommand(certUpdateCmd)
	certCmd.AddCommand(certApplyAWSCmd)
	certCmd.AddCommand(certValidationInfoCmd)

	// Register cert command with root
	rootCmd.AddCommand(certCmd)
}

func executeCertList() (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	rawResp, err := client.Get("/API/cdn/sslcert", nil)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			CommonName string `json:"common_name"`
			ExpireTime string `json:"expire_time"`
			State      string `json:"state"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if resp.Code != 1 {
		msg := resp.Message
		if msg == "" {
			msg = "the server returned an error with no details. Check that the resource exists and parameters are correct"
		}
		return "", fmt.Errorf("API error (code %d): %s", resp.Code, msg)
	}

	headers := []string{"ID", "NAME", "DOMAIN", "STATE", "EXPIRY"}
	rows := make([][]string, 0, len(resp.Data))
	for _, c := range resp.Data {
		rows = append(rows, []string{c.ID, c.Name, c.CommonName, c.State, c.ExpireTime})
	}
	return formatTable(headers, rows), nil
}

func executeCertUpload(name, certFile, keyFile string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	certContent, err := os.ReadFile(certFile)
	if err != nil {
		return "", fmt.Errorf("failed to read certificate file: %w", err)
	}
	keyContent, err := os.ReadFile(keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	body := map[string]string{
		"name":        name,
		"certificate": string(certContent),
		"private_key": string(keyContent),
	}
	rawResp, err := client.Post("/API/cdn/sslcert", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	return fmt.Sprintf("Certificate '%s' uploaded successfully.", name), nil
}

func executeCertUpdate(id, certFile, keyFile string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	certContent, err := os.ReadFile(certFile)
	if err != nil {
		return "", fmt.Errorf("failed to read certificate file: %w", err)
	}
	keyContent, err := os.ReadFile(keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	body := map[string]string{
		"id":          id,
		"certificate": string(certContent),
		"private_key": string(keyContent),
	}
	rawResp, err := client.Put("/API/cdn/sslcert", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	return fmt.Sprintf("Certificate '%s' updated successfully.", id), nil
}

func executeCertApplyAWS(domain string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := map[string]string{"domain": domain}
	rawResp, err := client.Post("/API/cdn/sslcert/apply", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	return fmt.Sprintf("AWS certificate application for '%s' submitted successfully.", domain), nil
}

func executeCertValidationInfo(id string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"id": id}
	rawResp, err := client.Get("/API/cdn/sslcert/validation/options", queryParams)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}
