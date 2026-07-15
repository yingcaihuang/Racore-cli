package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"racore-cli/internal/mcp"
	"racore-cli/internal/auth"
	"racore-cli/internal/credential"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server mode for AI agent communication",
	Example: `  # Start with credentials from ~/.racore/credentials (requires prior login)
  racore-cli serve

  # Start with credentials via flags
  racore-cli serve --access-key YOUR_KEY --secret-key YOUR_SECRET

  # Start with credentials via environment variables
  RACORE_ACCESS_KEY=key RACORE_SECRET_KEY=secret racore-cli serve

  # MCP config with env vars (recommended for MCP servers):
  # {
  #   "mcpServers": {
  #     "racore": {
  #       "command": "racore-cli",
  #       "args": ["serve"],
  #       "env": {
  #         "RACORE_ACCESS_KEY": "your_access_key",
  #         "RACORE_SECRET_KEY": "your_secret_key"
  #       }
  #     }
  #   }
  # }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If credentials passed via flags, set them as env vars for the session
		accessKey, _ := cmd.Flags().GetString("access-key")
		secretKey, _ := cmd.Flags().GetString("secret-key")
		if accessKey != "" && secretKey != "" {
			os.Setenv("RACORE_ACCESS_KEY", accessKey)
			os.Setenv("RACORE_SECRET_KEY", secretKey)
		}

		server := mcp.NewServer(os.Stdin, os.Stdout, os.Stderr)

		// Register login tool
		server.RegisterTool(mcp.ToolDefinition{
			Name:        "login",
			Description: "Authenticate with Racore Cloud using access key and secret key",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"access_key": {"type": "string", "description": "Racore Cloud access key"},
					"secret_key": {"type": "string", "description": "Racore Cloud secret key"}
				},
				"required": ["access_key", "secret_key"]
			}`),
		}, loginHandler)

		// Register whoami tool
		server.RegisterTool(mcp.ToolDefinition{
			Name:        "whoami",
			Description: "Display current authentication status",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		}, whoamiHandler)

		// Register logout tool
		server.RegisterTool(mcp.ToolDefinition{
			Name:        "logout",
			Description: "Clear stored credentials and revoke local access",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {}
			}`),
		}, logoutHandler)

		registerDomainTools(server)
		registerCacheTools(server)
		registerStatsTools(server)
		registerCertTools(server)
		registerWorkorderTools(server)
		registerLogTools(server)

		return server.Serve()
	},
}

func init() {
	serveCmd.Flags().String("access-key", "", "Racore Cloud access key (alternative to RACORE_ACCESS_KEY env var)")
	serveCmd.Flags().String("secret-key", "", "Racore Cloud secret key (alternative to RACORE_SECRET_KEY env var)")
	rootCmd.AddCommand(serveCmd)
}

// --- Domain tool registration ---

func registerDomainTools(server *mcp.Server) {
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "domain-list",
		Description: "List CDN domains managed by your Racore Cloud account",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"filter": {"type": "string", "description": "Filter domains by name"}
			}
		}`),
	}, domainListHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "domain-create",
		Description: "Create a new CDN domain with SSL certificate (required). Automatically matches a wildcard certificate from the cert list. If no matching cert is found, creation fails unless no_cert=true is explicitly set.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "Domain name to create"},
				"type": {"type": "string", "description": "CDN type: oversea, live, video/vod, dynamic, static, download"},
				"source": {"type": "string", "description": "Origin source address (IP or domain)"},
				"source_type": {"type": "string", "description": "Origin type: 1 (IP) or 2 (domain), default: 2"},
				"cert_id": {"type": "string", "description": "SSL certificate ID. If not provided, auto-matches wildcard cert. Fails if no match found."},
				"no_cert": {"type": "boolean", "description": "Skip SSL certificate binding (not recommended). Only set to true if you explicitly want no HTTPS."}
			},
			"required": ["domain", "type", "source"]
		}`),
	}, domainCreateHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "domain-delete",
		Description: "Delete a CDN domain",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "Domain name to delete"}
			},
			"required": ["domain"]
		}`),
	}, domainDeleteHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "domain-enable",
		Description: "Enable a CDN domain",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "Domain name to enable"}
			},
			"required": ["domain"]
		}`),
	}, domainEnableHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "domain-disable",
		Description: "Disable a CDN domain",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "Domain name to disable"}
			},
			"required": ["domain"]
		}`),
	}, domainDisableHandler)

	// Domain config get/set tools
	type domainConfigTool struct {
		name     string
		desc     string
		endpoint string
		method   string
	}

	configTools := []domainConfigTool{
		{"domain-source", "origin source", "/API/cdn/domain/source", "put"},
		{"domain-ssl", "SSL/HTTPS", "/API/cdn/domain/ssl", "put"},
		{"domain-enforce-https", "HTTPS enforcement", "/API/cdn/domain/enforce/https", "put"},
		{"domain-ip-filter", "IP filter", "/API/cdn/domain/ip/filter", "post"},
		{"domain-referer-filter", "referer filter", "/API/cdn/domain/referer/filter", "post"},
		{"domain-ua-filter", "user-agent filter", "/API/cdn/domain/user/agent/filter", "post"},
		{"domain-origin-protocol", "origin protocol policy", "/API/cdn/domain/origin/protocol/policy", "put"},
		{"domain-http2", "HTTP/2", "/API/cdn/domain/http2", "put"},
		{"domain-http3", "HTTP/3", "/API/cdn/domain/http3", "put"},
		{"domain-tls-version", "minimum TLS version", "/API/cdn/domain/min/tls/version", "put"},
		{"domain-compress", "page compression", "/API/cdn/domain/page/compress", "put"},
		{"domain-ipv6", "IPv6", "/API/cdn/domain/ipv6", "put"},
		{"domain-cache-policy", "cache policy", "/API/cdn/domain/cache/conf", "put"},
		{"domain-origin-host", "origin host", "/API/cdn/domain/origin/host", "put"},
		{"domain-origin-timeout", "origin connection timeout", "/API/cdn/domain/origin/connection/policy", "put"},
		{"domain-geo-restriction", "geographic restriction", "/API/cdn/domain/geo/restriction", "put"},
		{"domain-request-headers", "HTTP request headers", "/API/cdn/domain/http/request/headers", "post"},
		{"domain-response-headers", "HTTP response headers", "/API/cdn/domain/http/response/headers", "post"},
		{"domain-request-header-policy", "request header policy", "/API/cdn/domain/request/header/policy", "put"},
		{"domain-response-header-policy", "response header policy", "/API/cdn/domain/response/header/policy", "put"},
	}

	for _, t := range configTools {
		t := t // capture loop variable
		// Register GET tool
		server.RegisterTool(mcp.ToolDefinition{
			Name:        t.name + "-get",
			Description: fmt.Sprintf("Get %s configuration for a domain", t.desc),
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"domain": {"type": "string", "description": "Domain name"}
				},
				"required": ["domain"]
			}`),
		}, func(params json.RawMessage) (interface{}, error) {
			var input struct {
				Domain string `json:"domain"`
			}
			if err := json.Unmarshal(params, &input); err != nil {
				return nil, fmt.Errorf("invalid parameters: %w", err)
			}
			if input.Domain == "" {
				return makeErrorResult("domain is required"), nil
			}
			output, err := executeDomainConfigGet(t.endpoint, input.Domain)
			if err != nil {
				return makeErrorResult(err.Error()), nil
			}
			return makeTextResult(output), nil
		})

		// Register SET tool
		server.RegisterTool(mcp.ToolDefinition{
			Name:        t.name + "-set",
			Description: fmt.Sprintf("Set %s configuration for a domain", t.desc),
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"domain": {"type": "string", "description": "Domain name"},
					"config": {"type": "string", "description": "Configuration as JSON string"}
				},
				"required": ["domain", "config"]
			}`),
		}, func(params json.RawMessage) (interface{}, error) {
			var input struct {
				Domain string `json:"domain"`
				Config string `json:"config"`
			}
			if err := json.Unmarshal(params, &input); err != nil {
				return nil, fmt.Errorf("invalid parameters: %w", err)
			}
			if input.Domain == "" || input.Config == "" {
				return makeErrorResult("domain and config are required"), nil
			}
			output, err := executeDomainConfigSet(t.endpoint, input.Domain, input.Config, t.method)
			if err != nil {
				return makeErrorResult(err.Error()), nil
			}
			return makeTextResult(output), nil
		})
	}
}

// --- Cache tool registration ---

func registerCacheTools(server *mcp.Server) {
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-purge",
		Description: "Purge cached URLs from CDN edge nodes",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"urls": {"type": "array", "items": {"type": "string"}, "description": "List of URLs to purge"}
			},
			"required": ["urls"]
		}`),
	}, cachePurgeHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-purge-status",
		Description: "Check the status of a cache purge task",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"task_id": {"type": "string", "description": "Purge task ID"}
			},
			"required": ["task_id"]
		}`),
	}, cachePurgeStatusHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-prefetch",
		Description: "Prefetch URLs to CDN edge nodes",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"urls": {"type": "array", "items": {"type": "string"}, "description": "List of URLs to prefetch"},
				"region": {"type": "string", "description": "Target region for prefetch"},
				"country": {"type": "string", "description": "Target country for prefetch"}
			},
			"required": ["urls"]
		}`),
	}, cachePrefetchHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-prefetch-status",
		Description: "Check the status of a cache prefetch task",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"task_id": {"type": "string", "description": "Prefetch task ID"}
			},
			"required": ["task_id"]
		}`),
	}, cachePrefetchStatusHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-prewarm-regions",
		Description: "Get available prewarm regions for a URL",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"url": {"type": "string", "description": "URL to get prewarm regions for"}
			},
			"required": ["url"]
		}`),
	}, cachePrewarmRegionsHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-prewarm-pop",
		Description: "Get prewarm PoP nodes for a region",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"region": {"type": "string", "description": "Region to get PoP nodes for"}
			},
			"required": ["region"]
		}`),
	}, cachePrewarmPopHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-list-policies",
		Description: "List available AWS cache policies",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "CDN domain name"},
				"type": {"type": "string", "description": "Policy type: managed or custom (default: managed)"}
			},
			"required": ["domain"]
		}`),
	}, cacheListPoliciesHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-list-origin-request-policies",
		Description: "List available AWS origin request policies",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "CDN domain name"},
				"type": {"type": "string", "description": "Policy type: managed or custom (default: managed)"}
			},
			"required": ["domain"]
		}`),
	}, cacheListOriginRequestPoliciesHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cache-list-response-header-policies",
		Description: "List available AWS response header policies",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "CDN domain name"},
				"type": {"type": "string", "description": "Policy type: managed or custom (default: managed)"}
			},
			"required": ["domain"]
		}`),
	}, cacheListResponseHeaderPoliciesHandler)
}

// --- Stats tool registration ---

func registerStatsTools(server *mcp.Server) {
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-flow",
		Description: "Query CDN bandwidth/flow statistics",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time for the query"},
				"end_time": {"type": "string", "description": "End time for the query"},
				"domains": {"type": "array", "items": {"type": "string"}, "description": "List of domains to query"},
				"interval": {"type": "string", "description": "Time interval for aggregation"}
			},
			"required": ["start_time", "end_time"]
		}`),
	}, statsFlowHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-request",
		Description: "Query CDN request count statistics",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time for the query"},
				"end_time": {"type": "string", "description": "End time for the query"},
				"domains": {"type": "array", "items": {"type": "string"}, "description": "List of domains to query"},
				"interval": {"type": "string", "description": "Time interval for aggregation"}
			},
			"required": ["start_time", "end_time"]
		}`),
	}, statsRequestHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-hit-flow",
		Description: "Query CDN cache hit flow statistics",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time for the query"},
				"end_time": {"type": "string", "description": "End time for the query"},
				"domains": {"type": "array", "items": {"type": "string"}, "description": "List of domains to query"}
			},
			"required": ["start_time", "end_time"]
		}`),
	}, statsHitFlowHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-hit-request",
		Description: "Query CDN cache hit request statistics",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time for the query"},
				"end_time": {"type": "string", "description": "End time for the query"},
				"domains": {"type": "array", "items": {"type": "string"}, "description": "List of domains to query"}
			},
			"required": ["start_time", "end_time"]
		}`),
	}, statsHitRequestHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-http-code",
		Description: "Query CDN HTTP status code statistics",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time for the query"},
				"end_time": {"type": "string", "description": "End time for the query"},
				"domains": {"type": "array", "items": {"type": "string"}, "description": "List of domains to query"}
			},
			"required": ["start_time", "end_time"]
		}`),
	}, statsHTTPCodeHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-http-code-detail",
		Description: "Query CDN HTTP status code detail statistics",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time for the query"},
				"end_time": {"type": "string", "description": "End time for the query"},
				"domains": {"type": "array", "items": {"type": "string"}, "description": "List of domains to query"}
			},
			"required": ["start_time", "end_time"]
		}`),
	}, statsHTTPCodeDetailHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-district",
		Description: "Query CDN traffic distribution by country/region",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time for the query"},
				"end_time": {"type": "string", "description": "End time for the query"},
				"domains": {"type": "array", "items": {"type": "string"}, "description": "List of domains to query"}
			},
			"required": ["start_time", "end_time"]
		}`),
	}, statsDistrictHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-iso-country",
		Description: "List ISO country/region codes",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
	}, statsISOCountryHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-top-domain",
		Description: "Query top domains by traffic",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time for the query"},
				"end_time": {"type": "string", "description": "End time for the query"}
			},
			"required": ["start_time", "end_time"]
		}`),
	}, statsTopDomainHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-top-url",
		Description: "Query top URLs by traffic",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time (yyyy-mm-dd hh:mm)"},
				"end_time": {"type": "string", "description": "End time (yyyy-mm-dd hh:mm)"},
				"scope": {"type": "string", "description": "Time scope (today, yesterday, week, month, last_month)"},
				"sorted": {"type": "string", "description": "Sort order (url_size or url_count)"}
			}
		}`),
	}, statsTopURLHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-top-referer",
		Description: "Query top referers by traffic",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time (yyyy-mm-dd hh:mm)"},
				"end_time": {"type": "string", "description": "End time (yyyy-mm-dd hh:mm)"},
				"scope": {"type": "string", "description": "Time scope (today, yesterday, week, month, last_month)"},
				"sorted": {"type": "string", "description": "Sort order (referer_size or referer_count)"}
			}
		}`),
	}, statsTopRefererHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "stats-top-ua",
		Description: "Query top user agents by traffic",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"start_time": {"type": "string", "description": "Start time (yyyy-mm-dd hh:mm)"},
				"end_time": {"type": "string", "description": "End time (yyyy-mm-dd hh:mm)"},
				"scope": {"type": "string", "description": "Time scope (today, yesterday, week, month, last_month)"},
				"sorted": {"type": "string", "description": "Sort order (ua_size or ua_count)"}
			}
		}`),
	}, statsTopUAHandler)
}

// --- Cert tool registration ---

func registerCertTools(server *mcp.Server) {
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cert-list",
		Description: "List all SSL certificates",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
	}, certListHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cert-upload",
		Description: "Upload a new SSL certificate",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": {"type": "string", "description": "Certificate name"},
				"cert_file": {"type": "string", "description": "Path to certificate file"},
				"key_file": {"type": "string", "description": "Path to private key file"}
			},
			"required": ["name", "cert_file", "key_file"]
		}`),
	}, certUploadHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cert-update",
		Description: "Update an existing SSL certificate",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id": {"type": "string", "description": "Certificate ID"},
				"cert_file": {"type": "string", "description": "Path to certificate file"},
				"key_file": {"type": "string", "description": "Path to private key file"}
			},
			"required": ["id", "cert_file", "key_file"]
		}`),
	}, certUpdateHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cert-apply-aws",
		Description: "Apply for an AWS certificate for a domain",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "Domain to apply certificate for"}
			},
			"required": ["domain"]
		}`),
	}, certApplyAWSHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "cert-validation-info",
		Description: "Get certificate validation information",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id": {"type": "string", "description": "Certificate ID"}
			},
			"required": ["id"]
		}`),
	}, certValidationInfoHandler)
}

// --- Workorder tool registration ---

func registerWorkorderTools(server *mcp.Server) {
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-list",
		Description: "List all work orders",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
	}, workorderListHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-create",
		Description: "Create a new work order",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"title": {"type": "string", "description": "Work order title"},
				"description": {"type": "string", "description": "Work order description"},
				"type": {"type": "string", "description": "Work order type"}
			},
			"required": ["title", "description", "type"]
		}`),
	}, workorderCreateHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-reopen",
		Description: "Reopen a work order",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id": {"type": "string", "description": "Work order ID"}
			},
			"required": ["id"]
		}`),
	}, workorderReopenHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-delete",
		Description: "Delete a work order",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id": {"type": "string", "description": "Work order ID"}
			},
			"required": ["id"]
		}`),
	}, workorderDeleteHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-cancel",
		Description: "Cancel a work order",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id": {"type": "string", "description": "Work order ID"}
			},
			"required": ["id"]
		}`),
	}, workorderCancelHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-close",
		Description: "Close a work order",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id": {"type": "string", "description": "Work order ID"}
			},
			"required": ["id"]
		}`),
	}, workorderCloseHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-types",
		Description: "List work order types/categories",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {}
		}`),
	}, workorderTypesHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-log",
		Description: "View work order communication log",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id": {"type": "string", "description": "Work order ID"}
			},
			"required": ["id"]
		}`),
	}, workorderLogHandler)

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "workorder-send-message",
		Description: "Send a message to a work order",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"id": {"type": "string", "description": "Work order ID"},
				"message": {"type": "string", "description": "Message content"}
			},
			"required": ["id", "message"]
		}`),
	}, workorderSendMessageHandler)
}

// --- Log tool registration ---

func registerLogTools(server *mcp.Server) {
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "log-list",
		Description: "List CDN log files with download URLs",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"domain": {"type": "string", "description": "Domain name"},
				"start_date": {"type": "string", "description": "Start date for log query"},
				"end_date": {"type": "string", "description": "End date for log query"}
			},
			"required": ["domain"]
		}`),
	}, logListHandler)
}

// --- Helper to build text content result ---

func makeTextResult(text string) interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

// makeErrorResult returns a tool result indicating an error occurred.
// Per MCP spec, this is returned as a successful JSON-RPC response with isError=true.
func makeErrorResult(text string) interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
		"isError": true,
	}
}

// --- Auth handlers ---

func loginHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.AccessKey == "" || input.SecretKey == "" {
		return makeErrorResult("access_key and secret_key are required"), nil
	}
	output, err := executeLogin(input.AccessKey, input.SecretKey)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func whoamiHandler(params json.RawMessage) (interface{}, error) {
	output, err := executeWhoami()
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func logoutHandler(params json.RawMessage) (interface{}, error) {
	output, err := executeLogout()
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

// --- Domain handlers ---

func domainListHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Filter string `json:"filter"`
	}
	if params != nil {
		json.Unmarshal(params, &input)
	}
	output, err := executeDomainList(input.Filter)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func domainCreateHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain     string `json:"domain"`
		Type       string `json:"type"`
		Source     string `json:"source"`
		SourceType string `json:"source_type"`
		CertID     string `json:"cert_id"`
		NoCert     bool   `json:"no_cert"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" || input.Type == "" || input.Source == "" {
		return makeErrorResult("domain, type, and source are required"), nil
	}
	sourceType := input.SourceType
	if sourceType == "" {
		sourceType = "2"
	}
	output, err := executeDomainCreate(input.Domain, input.Type, input.Source, sourceType, input.CertID, input.NoCert)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func domainDeleteHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" {
		return makeErrorResult("domain is required"), nil
	}
	output, err := executeDomainDelete(input.Domain)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func domainEnableHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" {
		return makeErrorResult("domain is required"), nil
	}
	output, err := executeDomainEnable(input.Domain)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func domainDisableHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" {
		return makeErrorResult("domain is required"), nil
	}
	output, err := executeDomainDisable(input.Domain)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

// --- Cache handlers ---

func cachePurgeHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		URLs []string `json:"urls"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if len(input.URLs) == 0 {
		return makeErrorResult("urls is required"), nil
	}
	output, err := executeCachePurge(input.URLs)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func cachePurgeStatusHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.TaskID == "" {
		return makeErrorResult("task_id is required"), nil
	}
	output, err := executeCachePurgeStatus(input.TaskID)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func cachePrefetchHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		URLs    []string `json:"urls"`
		Region  string   `json:"region"`
		Country string   `json:"country"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if len(input.URLs) == 0 {
		return makeErrorResult("urls is required"), nil
	}
	output, err := executeCachePrefetch(input.URLs, input.Region, input.Country)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func cachePrefetchStatusHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.TaskID == "" {
		return makeErrorResult("task_id is required"), nil
	}
	output, err := executeCachePrefetchStatus(input.TaskID)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func cachePrewarmRegionsHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.URL == "" {
		return makeErrorResult("url is required"), nil
	}
	output, err := executeCachePrewarmRegions(input.URL)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func cachePrewarmPopHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Region string `json:"region"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Region == "" {
		return makeErrorResult("region is required"), nil
	}
	output, err := executeCachePrewarmPop(input.Region)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func cacheListPoliciesHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain string `json:"domain"`
		Type   string `json:"type"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" {
		return makeErrorResult("domain is required"), nil
	}
	policyType := input.Type
	if policyType == "" {
		policyType = "managed"
	}
	output, err := executeCacheListPolicies(input.Domain, policyType)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func cacheListOriginRequestPoliciesHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain string `json:"domain"`
		Type   string `json:"type"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" {
		return makeErrorResult("domain is required"), nil
	}
	policyType := input.Type
	if policyType == "" {
		policyType = "managed"
	}
	output, err := executeCacheListOriginRequestPolicies(input.Domain, policyType)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func cacheListResponseHeaderPoliciesHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain string `json:"domain"`
		Type   string `json:"type"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" {
		return makeErrorResult("domain is required"), nil
	}
	policyType := input.Type
	if policyType == "" {
		policyType = "managed"
	}
	output, err := executeCacheListResponseHeaderPolicies(input.Domain, policyType)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

// --- Stats handlers ---

func statsFlowHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string   `json:"start_time"`
		EndTime   string   `json:"end_time"`
		Domains   []string `json:"domains"`
		Interval  string   `json:"interval"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsFlow(input.StartTime, input.EndTime, input.Domains, input.Interval)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsRequestHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string   `json:"start_time"`
		EndTime   string   `json:"end_time"`
		Domains   []string `json:"domains"`
		Interval  string   `json:"interval"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsRequest(input.StartTime, input.EndTime, input.Domains, input.Interval)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsHitFlowHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string   `json:"start_time"`
		EndTime   string   `json:"end_time"`
		Domains   []string `json:"domains"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsHitFlow(input.StartTime, input.EndTime, input.Domains)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsHitRequestHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string   `json:"start_time"`
		EndTime   string   `json:"end_time"`
		Domains   []string `json:"domains"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsHitRequest(input.StartTime, input.EndTime, input.Domains)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsHTTPCodeHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string   `json:"start_time"`
		EndTime   string   `json:"end_time"`
		Domains   []string `json:"domains"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsHTTPCode(input.StartTime, input.EndTime, input.Domains)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsHTTPCodeDetailHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string   `json:"start_time"`
		EndTime   string   `json:"end_time"`
		Domains   []string `json:"domains"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsHTTPCodeDetail(input.StartTime, input.EndTime, input.Domains)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsDistrictHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string   `json:"start_time"`
		EndTime   string   `json:"end_time"`
		Domains   []string `json:"domains"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsDistrict(input.StartTime, input.EndTime, input.Domains)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsISOCountryHandler(params json.RawMessage) (interface{}, error) {
	output, err := executeStatsISOCountry()
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsTopDomainHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsTopDomain(input.StartTime, input.EndTime)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsTopURLHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Scope     string `json:"scope"`
		Sorted    string `json:"sorted"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsTopURL(input.StartTime, input.EndTime, input.Scope, input.Sorted)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsTopRefererHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Scope     string `json:"scope"`
		Sorted    string `json:"sorted"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsTopReferer(input.StartTime, input.EndTime, input.Scope, input.Sorted)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func statsTopUAHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Scope     string `json:"scope"`
		Sorted    string `json:"sorted"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	output, err := executeStatsTopUA(input.StartTime, input.EndTime, input.Scope, input.Sorted)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

// --- Cert handlers ---

func certListHandler(params json.RawMessage) (interface{}, error) {
	output, err := executeCertList()
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func certUploadHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Name     string `json:"name"`
		CertFile string `json:"cert_file"`
		KeyFile  string `json:"key_file"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Name == "" || input.CertFile == "" || input.KeyFile == "" {
		return makeErrorResult("name, cert_file, and key_file are required"), nil
	}
	output, err := executeCertUpload(input.Name, input.CertFile, input.KeyFile)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func certUpdateHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		ID       string `json:"id"`
		CertFile string `json:"cert_file"`
		KeyFile  string `json:"key_file"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.ID == "" || input.CertFile == "" || input.KeyFile == "" {
		return makeErrorResult("id, cert_file, and key_file are required"), nil
	}
	output, err := executeCertUpdate(input.ID, input.CertFile, input.KeyFile)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func certApplyAWSHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" {
		return makeErrorResult("domain is required"), nil
	}
	output, err := executeCertApplyAWS(input.Domain)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func certValidationInfoHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.ID == "" {
		return makeErrorResult("id is required"), nil
	}
	output, err := executeCertValidationInfo(input.ID)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

// --- Workorder handlers ---

func workorderListHandler(params json.RawMessage) (interface{}, error) {
	output, err := executeWorkorderList()
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func workorderCreateHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Title == "" || input.Description == "" || input.Type == "" {
		return makeErrorResult("title, description, and type are required"), nil
	}
	output, err := executeWorkorderCreate(input.Title, input.Description, input.Type)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func workorderReopenHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.ID == "" {
		return makeErrorResult("id is required"), nil
	}
	output, err := executeWorkorderReopen(input.ID)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func workorderDeleteHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.ID == "" {
		return makeErrorResult("id is required"), nil
	}
	output, err := executeWorkorderDelete(input.ID)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func workorderCancelHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.ID == "" {
		return makeErrorResult("id is required"), nil
	}
	output, err := executeWorkorderCancel(input.ID)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func workorderCloseHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.ID == "" {
		return makeErrorResult("id is required"), nil
	}
	output, err := executeWorkorderClose(input.ID)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func workorderTypesHandler(params json.RawMessage) (interface{}, error) {
	output, err := executeWorkorderTypes()
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func workorderLogHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.ID == "" {
		return makeErrorResult("id is required"), nil
	}
	output, err := executeWorkorderLog(input.ID)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

func workorderSendMessageHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.ID == "" || input.Message == "" {
		return makeErrorResult("id and message are required"), nil
	}
	output, err := executeWorkorderSendMessage(input.ID, input.Message)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

// --- Log handlers ---

func logListHandler(params json.RawMessage) (interface{}, error) {
	var input struct {
		Domain    string `json:"domain"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	if input.Domain == "" {
		return makeErrorResult("domain is required"), nil
	}
	output, err := executeLogList(input.Domain, input.StartDate, input.EndDate)
	if err != nil {
		return makeErrorResult(err.Error()), nil
	}
	return makeTextResult(output), nil
}

// --- Execute functions for auth (login, whoami, logout) ---

func executeLogin(accessKey, secretKey string) (string, error) {
	creds := &credential.Credentials{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
	defer credential.ClearSensitive(creds)

	mgr := auth.NewManager(accessKey, secretKey)
	resp, err := mgr.Authenticate()
	if err != nil {
		return "", fmt.Errorf("authentication request failed: %w", err)
	}

	if resp.Code != 1 {
		msg := resp.Message
		if msg == "" {
			msg = "the server returned an error with no details. Check that the credentials are correct"
		}
		return "", fmt.Errorf("authentication failed (code %d): %s", resp.Code, msg)
	}

	creds.Token = resp.Data.Token
	creds.Expire = resp.Data.Expire

	store, err := credential.NewStore()
	if err != nil {
		return "", fmt.Errorf("cannot initialize credential store: %w", err)
	}

	if err := store.Save(creds); err != nil {
		return "", fmt.Errorf("cannot save credentials: %w", err)
	}

	expireTime := time.Unix(resp.Data.Expire, 0).Format(time.RFC3339)
	return fmt.Sprintf("Login successful. Token expires at %s", expireTime), nil
}

func executeWhoami() (string, error) {
	store, err := credential.NewStore()
	if err != nil {
		return "", fmt.Errorf("cannot initialize credential store: %w", err)
	}

	creds, err := store.Load()
	if err != nil {
		return "", fmt.Errorf("cannot load credentials: %w", err)
	}

	if creds == nil {
		return "", fmt.Errorf("not logged in")
	}

	var buf strings.Builder

	if warning, err := store.CheckPermissions(); err == nil && warning != "" {
		buf.WriteString(warning + "\n")
	}

	buf.WriteString(fmt.Sprintf("Access Key: %s\n", credential.MaskAccessKey(creds.AccessKey)))
	buf.WriteString(fmt.Sprintf("Storage: %s\n", store.StorageType()))

	remaining := creds.Expire - time.Now().Unix()
	if remaining >= 300 {
		hours := remaining / 3600
		minutes := (remaining % 3600) / 60
		buf.WriteString(fmt.Sprintf("Token Status: Valid (%dh %dm remaining)\n", hours, minutes))
	} else {
		buf.WriteString("Token Status: Expired/Expiring\n")
	}

	return buf.String(), nil
}

func executeLogout() (string, error) {
	store, err := credential.NewStore()
	if err != nil {
		return "", fmt.Errorf("cannot initialize credential store: %w", err)
	}

	creds, err := store.Load()
	if err != nil {
		return "", fmt.Errorf("cannot load credentials: %w", err)
	}

	if creds == nil {
		return "Already logged out", nil
	}

	if err := store.Delete(); err != nil {
		return "", fmt.Errorf("cannot delete credentials: %w", err)
	}

	return "Logged out successfully.", nil
}
