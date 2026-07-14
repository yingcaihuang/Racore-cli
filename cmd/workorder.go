package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var workorderCmd = &cobra.Command{
	Use:   "workorder",
	Short: "Work order management",
	Example: `  # List all work orders
  racore-cli workorder list

  # List work order types
  racore-cli workorder types

  # Create a work order
  racore-cli workorder create --title "Issue" --description "Details" --type "技术支持"

  # View communication log
  racore-cli workorder log --id 123

  # Send message
  racore-cli workorder send-message --id 123 --message "Hello"`,
}

var workorderListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all work orders",
	Example: `  racore-cli workorder list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := executeWorkorderList()
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

var workorderCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new work order",
	Example: `  racore-cli workorder create --title "Issue" --description "Details" --type "技术支持"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		woType, _ := cmd.Flags().GetString("type")
		output, err := executeWorkorderCreate(title, description, woType)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

var workorderReopenCmd = &cobra.Command{
	Use:     "reopen",
	Short:   "Reopen a work order",
	Example: `  racore-cli workorder reopen --id 123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		output, err := executeWorkorderReopen(id)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

var workorderDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a work order",
	Example: `  racore-cli workorder delete --id 123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		output, err := executeWorkorderDelete(id)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

var workorderCancelCmd = &cobra.Command{
	Use:     "cancel",
	Short:   "Cancel a work order",
	Example: `  racore-cli workorder cancel --id 123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		output, err := executeWorkorderCancel(id)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

var workorderCloseCmd = &cobra.Command{
	Use:     "close",
	Short:   "Close a work order",
	Example: `  racore-cli workorder close --id 123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		output, err := executeWorkorderClose(id)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

var workorderTypesCmd = &cobra.Command{
	Use:     "types",
	Short:   "List work order types/categories",
	Example: `  racore-cli workorder types`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := executeWorkorderTypes()
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

var workorderLogCmd = &cobra.Command{
	Use:     "log",
	Short:   "View work order log",
	Example: `  racore-cli workorder log --id 123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		output, err := executeWorkorderLog(id)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

var workorderSendMessageCmd = &cobra.Command{
	Use:   "send-message",
	Short: "Send a message to a work order",
	Example: `  racore-cli workorder send-message --id 123 --message "Hello"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		message, _ := cmd.Flags().GetString("message")
		output, err := executeWorkorderSendMessage(id, message)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	},
}

func init() {
	// Register flags
	workorderCreateCmd.Flags().String("title", "", "Work order title (required)")
	workorderCreateCmd.Flags().String("description", "", "Detailed description of the issue")
	workorderCreateCmd.Flags().String("type", "", "Work order type (use 'workorder types' to list available)")

	workorderReopenCmd.Flags().String("id", "", "Work order ID to reopen")
	workorderDeleteCmd.Flags().String("id", "", "Work order ID to delete")
	workorderCancelCmd.Flags().String("id", "", "Work order ID to cancel")
	workorderCloseCmd.Flags().String("id", "", "Work order ID to close")
	workorderLogCmd.Flags().String("id", "", "Work order ID to view log for")

	workorderSendMessageCmd.Flags().String("id", "", "Work order ID to send message to")
	workorderSendMessageCmd.Flags().String("message", "", "Message content to send")

	// Register subcommands
	workorderCmd.AddCommand(workorderListCmd)
	workorderCmd.AddCommand(workorderCreateCmd)
	workorderCmd.AddCommand(workorderReopenCmd)
	workorderCmd.AddCommand(workorderDeleteCmd)
	workorderCmd.AddCommand(workorderCancelCmd)
	workorderCmd.AddCommand(workorderCloseCmd)
	workorderCmd.AddCommand(workorderTypesCmd)
	workorderCmd.AddCommand(workorderLogCmd)
	workorderCmd.AddCommand(workorderSendMessageCmd)

	// Register parent command
	rootCmd.AddCommand(workorderCmd)
}

func executeWorkorderList() (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	rawResp, err := client.Get("/API/user/workorder", nil)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Status    string `json:"status"`
			CreatedAt string `json:"created_at"`
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
	headers := []string{"ID", "TITLE", "STATUS", "CREATED_AT"}
	rows := make([][]string, 0, len(resp.Data))
	for _, w := range resp.Data {
		rows = append(rows, []string{w.ID, w.Title, w.Status, w.CreatedAt})
	}
	return formatTable(headers, rows), nil
}

func executeWorkorderCreate(title, description, woType string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	body := map[string]string{"title": title, "description": description, "type": woType}
	rawResp, err := client.Post("/API/user/workorder", body)
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
	return fmt.Sprintf("Work order '%s' created successfully.", title), nil
}

func executeWorkorderReopen(id string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	body := map[string]string{"id": id}
	rawResp, err := client.Put("/API/user/workorder", body)
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
	return fmt.Sprintf("Work order '%s' reopened successfully.", id), nil
}

func executeWorkorderDelete(id string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	body := map[string]string{"id": id}
	rawResp, err := client.Delete("/API/user/workorder", body)
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
	return fmt.Sprintf("Work order '%s' deleted successfully.", id), nil
}

func executeWorkorderCancel(id string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	body := map[string]string{"id": id}
	rawResp, err := client.Put("/API/user/workorder/cancel", body)
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
	return fmt.Sprintf("Work order '%s' cancelled successfully.", id), nil
}

func executeWorkorderClose(id string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	body := map[string]string{"id": id}
	rawResp, err := client.Put("/API/user/workorder/close", body)
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
	return fmt.Sprintf("Work order '%s' closed successfully.", id), nil
}

func executeWorkorderTypes() (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	rawResp, err := client.Get("/API/user/workorder/category", nil)
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

func executeWorkorderLog(id string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	queryParams := map[string]string{"id": id}
	rawResp, err := client.Get("/API/user/workorder/log", queryParams)
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

func executeWorkorderSendMessage(id, message string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}
	body := map[string]string{"id": id, "message": message}
	rawResp, err := client.Post("/API/user/workorder/log", body)
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
	return fmt.Sprintf("Message sent to work order '%s' successfully.", id), nil
}
