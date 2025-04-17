package fiberdocs

import (
	"embed"
	// "encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
)

//go:embed templates
var templatesFS embed.FS

// Config holds the configuration for the Fiber Docs middleware
type Config struct {
	// Title of the documentation page
	Title string
	// Description of the API
	Description string
	// Version of the API
	Version string
	// BasePath for the API
	BasePath string
	// UIPath is the path where the docs UI will be served
	UIPath string
	// RouteGrouping determines how to organize routes in the UI
	// Options: "path" (default), "method", "none"
	RouteGrouping string
	// AdditionalInfo holds extra information for routes that's not available from GetRoutes()
	// Map key is route path, value is map of properties
	AdditionalInfo map[string]map[string]interface{}
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Title:         "API Documentation",
		Description:   "Generated API documentation",
		Version:       "1.0.0",
		BasePath:      "",
		UIPath:        "/docs",
		RouteGrouping: "path",
		AdditionalInfo: make(map[string]map[string]interface{}),
	}
}

// Route represents a route with additional documentation info
type Route struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Params      []string               `json:"params"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Group       string                 `json:"group"`
	RequestBody interface{}            `json:"requestBody"`
	Responses   map[string]interface{} `json:"responses"`
}

// NewMiddleware creates a new Fiber Docs middleware
func NewMiddleware(config ...Config) fiber.Handler {
	// Set default config
	cfg := DefaultConfig()
	
	// Override with user config if provided
	if len(config) > 0 {
		cfg = config[0]
	}
	
	// Parse templates
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse templates: %v", err))
	}
	
	return func(c *fiber.Ctx) error {
		app := c.App()
		
		// Only serve on the configured UI path
		if c.Path() != cfg.UIPath && !strings.HasPrefix(c.Path(), "/api/v2/"+cfg.UIPath+"/") {
			return c.Next()
		}
		
		// If requesting JSON data
		if c.Path() == cfg.UIPath+"/data.json" {
			routes := processRoutes(app.GetRoutes(), cfg)
			return c.JSON(routes)
		}
		
		// If requesting the UI - serve the HTML
		var data struct {
			Config Config
		}
		data.Config = cfg
		
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		buf := new(strings.Builder)
		if err := tmpl.ExecuteTemplate(buf, "index.html", data); err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Template error")
		}
		
		return c.SendString(buf.String())
	}
}

// processRoutes transforms Fiber routes into document-friendly format
func processRoutes(fiberRoutes []fiber.Route, cfg Config) []Route {
	routes := make([]Route, 0, len(fiberRoutes))
	
	for _, r := range fiberRoutes {
		// Extract path parameters from path
		params := extractParams(r.Path)
		
		route := Route{
			Method: r.Method,
			Path:   r.Path,
			Name:   r.Name,
			Params: params,
			Group:  determineGroup(r.Path, r.Method, cfg.RouteGrouping),
			Responses: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
				},
			},
		}
		
		// Add additional info if available
		if info, exists := cfg.AdditionalInfo[r.Path]; exists {
			if desc, ok := info["description"].(string); ok {
				route.Description = desc
			}
			if reqBody, ok := info["requestBody"]; ok {
				route.RequestBody = reqBody
			}
			if responses, ok := info["responses"].(map[string]interface{}); ok {
				route.Responses = responses
			}
		}
		
		routes = append(routes, route)
	}
	
	// Sort routes by group then by path
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Group == routes[j].Group {
			return routes[i].Path < routes[j].Path
		}
		return routes[i].Group < routes[j].Group
	})
	
	return routes
}

// extractParams extracts path parameters from a route path
func extractParams(path string) []string {
	params := []string{}
	segments := strings.Split(path, "/")
	
	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			// Remove the colon
			params = append(params, segment[1:])
		}
	}
	
	return params
}

// determineGroup determines the group for a route based on grouping strategy
func determineGroup(path, method, groupingStrategy string) string {
	switch groupingStrategy {
	case "method":
		return method
	case "path":
		segments := strings.Split(path, "/")
		if len(segments) > 1 {
			return segments[1] // First non-empty segment
		}
		return "default"
	default:
		return "default"
	}
}

// AddRouteInfo adds additional documentation for a specific route
// This allows adding descriptions, request/response examples, etc.
func (cfg *Config) AddRouteInfo(path string, info map[string]interface{}) {
	if cfg.AdditionalInfo == nil {
		cfg.AdditionalInfo = make(map[string]map[string]interface{})
	}
	cfg.AdditionalInfo[path] = info
}

// New sets up the docs middleware with the given config
func New(config Config) fiber.Handler {
	return NewMiddleware(config)
}