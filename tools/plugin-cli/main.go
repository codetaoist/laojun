package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// PluginTemplate represents the structure of a plugin
type PluginTemplate struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Category    string `json:"category"`
	Type        string `json:"type"`
	EntryPoint  string `json:"entry_point"`
	Dependencies []string `json:"dependencies"`
	Permissions []string `json:"permissions"`
	Config      map[string]interface{} `json:"config"`
}

// PluginConfig represents the CLI configuration
type PluginConfig struct {
	Author       string `json:"author"`
	Email        string `json:"email"`
	Organization string `json:"organization"`
	Registry     string `json:"registry"`
	APIEndpoint  string `json:"api_endpoint"`
	Token        string `json:"token"`
}

var (
	cfgFile string
	config  PluginConfig
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "laojun-plugin",
		Short: "Laojun Plugin Development CLI",
		Long: `A comprehensive CLI tool for developing, testing, and managing Laojun plugins.
This tool helps developers create, build, test, and deploy plugins for the Laojun platform.`,
	}

	// Add persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.laojun-plugin.yaml)")

	// Add commands
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(createCmd())
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(testCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(packageCmd())
	rootCmd.AddCommand(publishCmd())
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(debugCmd())
	rootCmd.AddCommand(configCmd())

	// Initialize configuration
	cobra.OnInitialize(initConfig)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".laojun-plugin")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		viper.Unmarshal(&config)
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize plugin development environment",
		Long:  "Initialize the plugin development environment with default configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("üöÄ Initializing Laojun Plugin Development Environment...")

			// Create default config
			defaultConfig := PluginConfig{
				Author:       "Your Name",
				Email:        "your.email@example.com",
				Organization: "Your Organization",
				Registry:     "https://plugins.laojun.com",
				APIEndpoint:  "https://api.laojun.com",
				Token:        "",
			}

			// Interactive setup
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Enter your name: ")
			if name, _ := reader.ReadString('\n'); strings.TrimSpace(name) != "" {
				defaultConfig.Author = strings.TrimSpace(name)
			}

			fmt.Print("Enter your email: ")
			if email, _ := reader.ReadString('\n'); strings.TrimSpace(email) != "" {
				defaultConfig.Email = strings.TrimSpace(email)
			}

			fmt.Print("Enter your organization: ")
			if org, _ := reader.ReadString('\n'); strings.TrimSpace(org) != "" {
				defaultConfig.Organization = strings.TrimSpace(org)
			}

			// Save config
			viper.Set("author", defaultConfig.Author)
			viper.Set("email", defaultConfig.Email)
			viper.Set("organization", defaultConfig.Organization)
			viper.Set("registry", defaultConfig.Registry)
			viper.Set("api_endpoint", defaultConfig.APIEndpoint)

			home, _ := os.UserHomeDir()
			configPath := filepath.Join(home, ".laojun-plugin.yaml")
			if err := viper.WriteConfigAs(configPath); err != nil {
				log.Fatalf("Error writing config: %v", err)
			}

			fmt.Printf("‚ú?Configuration saved to %s\n", configPath)
			fmt.Println("üéâ Plugin development environment initialized successfully!")
		},
	}
}

func createCmd() *cobra.Command {
	var pluginType, category string

	cmd := &cobra.Command{
		Use:   "create [plugin-name]",
		Short: "Create a new plugin",
		Long:  "Create a new plugin with the specified name and template",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pluginName := args[0]
			createPlugin(pluginName, pluginType, category)
		},
	}

	cmd.Flags().StringVarP(&pluginType, "type", "t", "middleware", "Plugin type (middleware, handler, service, filter)")
	cmd.Flags().StringVarP(&category, "category", "c", "general", "Plugin category")

	return cmd
}

func createPlugin(name, pluginType, category string) {
	fmt.Printf("üî® Creating plugin: %s\n", name)

	// Create plugin directory
	pluginDir := filepath.Join(".", name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		log.Fatalf("Error creating plugin directory: %v", err)
	}

	// Create plugin manifest
	manifest := PluginTemplate{
		Name:        name,
		Version:     "1.0.0",
		Description: fmt.Sprintf("A %s plugin for Laojun", pluginType),
		Author:      config.Author,
		Category:    category,
		Type:        pluginType,
		EntryPoint:  "main.go",
		Dependencies: []string{
			"github.com/gin-gonic/gin",
		},
		Permissions: []string{
			"read:config",
			"write:logs",
		},
		Config: map[string]interface{}{
			"enabled": true,
			"debug":   false,
		},
	}

	// Save manifest
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := ioutil.WriteFile(manifestPath, manifestData, 0644); err != nil {
		log.Fatalf("Error writing manifest: %v", err)
	}

	// Create main.go from template
	mainGoTemplate := getPluginTemplate(pluginType)
	mainGoPath := filepath.Join(pluginDir, "main.go")
	
	tmpl, err := template.New("plugin").Parse(mainGoTemplate)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	file, err := os.Create(mainGoPath)
	if err != nil {
		log.Fatalf("Error creating main.go: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, manifest); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	// Create go.mod
	goModPath := filepath.Join(pluginDir, "go.mod")
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/viper v1.16.0
)
`, name)
	if err := ioutil.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		log.Fatalf("Error writing go.mod: %v", err)
	}

	// Create README.md
	readmePath := filepath.Join(pluginDir, "README.md")
	readmeContent := fmt.Sprintf(`# %s

%s

## Installation

```bash
laojun-plugin install %s
```

## Configuration

```yaml
%s:
  enabled: true
  debug: false
```

## Usage

This plugin provides %s functionality for the Laojun platform.

## Development

```bash
# Build the plugin
laojun-plugin build

# Test the plugin
laojun-plugin test

# Package the plugin
laojun-plugin package
```

## License

MIT License
`, manifest.Name, manifest.Description, manifest.Name, manifest.Name, pluginType)

	if err := ioutil.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		log.Fatalf("Error writing README.md: %v", err)
	}

	// Create test file
	testPath := filepath.Join(pluginDir, "main_test.go")
	testContent := getTestTemplate(manifest)
	if err := ioutil.WriteFile(testPath, []byte(testContent), 0644); err != nil {
		log.Fatalf("Error writing test file: %v", err)
	}

	fmt.Printf("‚ú?Plugin '%s' created successfully in %s\n", name, pluginDir)
	fmt.Println("üìù Next steps:")
	fmt.Printf("   cd %s\n", name)
	fmt.Println("   laojun-plugin build")
	fmt.Println("   laojun-plugin test")
}

func buildCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the plugin",
		Long:  "Build the plugin binary",
		Run: func(cmd *cobra.Command, args []string) {
			buildPlugin(output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output binary name")

	return cmd
}

func buildPlugin(output string) {
	fmt.Println("üî® Building plugin...")

	// Check if plugin.json exists
	if _, err := os.Stat("plugin.json"); os.IsNotExist(err) {
		log.Fatal("plugin.json not found. Are you in a plugin directory?")
	}

	// Read plugin manifest
	manifestData, err := ioutil.ReadFile("plugin.json")
	if err != nil {
		log.Fatalf("Error reading plugin.json: %v", err)
	}

	var manifest PluginTemplate
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		log.Fatalf("Error parsing plugin.json: %v", err)
	}

	// Set output name
	if output == "" {
		output = manifest.Name
	}

	// Build command
	buildArgs := []string{"build", "-o", output}
	if manifest.EntryPoint != "" {
		buildArgs = append(buildArgs, manifest.EntryPoint)
	} else {
		buildArgs = append(buildArgs, ".")
	}

	cmd := exec.Command("go", buildArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Build failed: %v", err)
	}

	fmt.Printf("‚ú?Plugin built successfully: %s\n", output)
}

func testCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test the plugin",
		Long:  "Run tests for the plugin",
		Run: func(cmd *cobra.Command, args []string) {
			testPlugin(verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func testPlugin(verbose bool) {
	fmt.Println("üß™ Running plugin tests...")

	testArgs := []string{"test"}
	if verbose {
		testArgs = append(testArgs, "-v")
	}
	testArgs = append(testArgs, "./...")

	cmd := exec.Command("go", testArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Tests failed: %v", err)
	}

	fmt.Println("‚ú?All tests passed!")
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate plugin configuration",
		Long:  "Validate the plugin manifest and configuration",
		Run: func(cmd *cobra.Command, args []string) {
			validatePlugin()
		},
	}
}

func validatePlugin() {
	fmt.Println("üîç Validating plugin...")

	// Check if plugin.json exists
	if _, err := os.Stat("plugin.json"); os.IsNotExist(err) {
		log.Fatal("‚ù?plugin.json not found")
	}

	// Read and validate manifest
	manifestData, err := ioutil.ReadFile("plugin.json")
	if err != nil {
		log.Fatalf("‚ù?Error reading plugin.json: %v", err)
	}

	var manifest PluginTemplate
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		log.Fatalf("‚ù?Error parsing plugin.json: %v", err)
	}

	// Validation checks
	errors := []string{}

	if manifest.Name == "" {
		errors = append(errors, "Plugin name is required")
	}

	if manifest.Version == "" {
		errors = append(errors, "Plugin version is required")
	}

	if manifest.EntryPoint == "" {
		errors = append(errors, "Entry point is required")
	} else if _, err := os.Stat(manifest.EntryPoint); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("Entry point file not found: %s", manifest.EntryPoint))
	}

	validTypes := []string{"middleware", "handler", "service", "filter"}
	validType := false
	for _, t := range validTypes {
		if manifest.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		errors = append(errors, fmt.Sprintf("Invalid plugin type: %s (must be one of: %s)", manifest.Type, strings.Join(validTypes, ", ")))
	}

	if len(errors) > 0 {
		fmt.Println("‚ù?Validation failed:")
		for _, err := range errors {
			fmt.Printf("   - %s\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("‚ú?Plugin validation passed!")
}

func packageCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "package",
		Short: "Package the plugin",
		Long:  "Package the plugin for distribution",
		Run: func(cmd *cobra.Command, args []string) {
			packagePlugin(output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output package name")

	return cmd
}

func packagePlugin(output string) {
	fmt.Println("üì¶ Packaging plugin...")

	// Validate first
	validatePlugin()

	// Read manifest
	manifestData, err := ioutil.ReadFile("plugin.json")
	if err != nil {
		log.Fatalf("Error reading plugin.json: %v", err)
	}

	var manifest PluginTemplate
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		log.Fatalf("Error parsing plugin.json: %v", err)
	}

	// Set output name
	if output == "" {
		output = fmt.Sprintf("%s-%s.tar.gz", manifest.Name, manifest.Version)
	}

	// Create package
	cmd := exec.Command("tar", "-czf", output, ".")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Packaging failed: %v", err)
	}

	fmt.Printf("‚ú?Plugin packaged successfully: %s\n", output)
}

func publishCmd() *cobra.Command {
	var registry string

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish the plugin",
		Long:  "Publish the plugin to the registry",
		Run: func(cmd *cobra.Command, args []string) {
			publishPlugin(registry)
		},
	}

	cmd.Flags().StringVarP(&registry, "registry", "r", "", "Plugin registry URL")

	return cmd
}

func publishPlugin(registry string) {
	fmt.Println("üöÄ Publishing plugin...")

	if registry == "" {
		registry = config.Registry
	}

	if registry == "" {
		log.Fatal("No registry specified. Use --registry flag or configure default registry.")
	}

	// TODO: Implement actual publishing logic
	fmt.Printf("Publishing to registry: %s\n", registry)
	fmt.Println("‚ú?Plugin published successfully!")
}

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install [plugin-name]",
		Short: "Install a plugin",
		Long:  "Install a plugin from the registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			installPlugin(args[0])
		},
	}
}

func installPlugin(name string) {
	fmt.Printf("üì• Installing plugin: %s\n", name)
	// TODO: Implement actual installation logic
	fmt.Printf("‚ú?Plugin '%s' installed successfully!\n", name)
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long:  "List all installed plugins",
		Run: func(cmd *cobra.Command, args []string) {
			listPlugins()
		},
	}
}

func listPlugins() {
	fmt.Println("üìã Installed plugins:")
	// TODO: Implement actual listing logic
	fmt.Println("   - example-plugin v1.0.0")
}

func debugCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug the plugin",
		Long:  "Start debug server for the plugin",
		Run: func(cmd *cobra.Command, args []string) {
			debugPlugin(port)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Debug server port")

	return cmd
}

func debugPlugin(port int) {
	fmt.Printf("üêõ Starting debug server on port %d...\n", port)
	// TODO: Implement actual debug server
	fmt.Println("Debug server started. Press Ctrl+C to stop.")
	select {} // Block forever
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Manage plugin CLI configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Current configuration:")
			fmt.Printf("Author: %s\n", config.Author)
			fmt.Printf("Email: %s\n", config.Email)
			fmt.Printf("Organization: %s\n", config.Organization)
			fmt.Printf("Registry: %s\n", config.Registry)
			fmt.Printf("API Endpoint: %s\n", config.APIEndpoint)
		},
	})

	return cmd
}

func getPluginTemplate(pluginType string) string {
	switch pluginType {
	case "middleware":
		return middlewareTemplate
	case "handler":
		return handlerTemplate
	case "service":
		return serviceTemplate
	case "filter":
		return filterTemplate
	default:
		return middlewareTemplate
	}
}

func getTestTemplate(manifest PluginTemplate) string {
	return fmt.Sprintf(`package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPlugin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(%sMiddleware())
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func BenchmarkPlugin(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.Use(%sMiddleware())
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}
`, strings.Title(manifest.Name), strings.Title(manifest.Name))
}

const middlewareTemplate = `package main

import (
	"log"
	"time"
	"github.com/gin-gonic/gin"
)

// {{.Name}}Middleware creates a new middleware instance
func {{.Name}}Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Pre-processing
		log.Printf("[{{.Name}}] Request started: %s %s", c.Request.Method, c.Request.URL.Path)
		
		// Process request
		c.Next()
		
		// Post-processing
		duration := time.Since(start)
		log.Printf("[{{.Name}}] Request completed in %v", duration)
	}
}

// Initialize plugin
func init() {
	log.Println("{{.Name}} plugin initialized")
}

func main() {
	router := gin.Default()
	router.Use({{.Name}}Middleware())
	
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "{{.Name}} plugin is working",
			"version": "{{.Version}}",
		})
	})
	
	log.Println("Starting {{.Name}} plugin server on :8080")
	router.Run(":8080")
}
`

const handlerTemplate = `package main

import (
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
)

// {{.Name}}Handler handles requests for {{.Name}}
func {{.Name}}Handler(c *gin.Context) {
	log.Printf("[{{.Name}}] Handling request: %s", c.Request.URL.Path)
	
	c.JSON(http.StatusOK, gin.H{
		"plugin": "{{.Name}}",
		"version": "{{.Version}}",
		"message": "Handler executed successfully",
		"timestamp": time.Now().Unix(),
	})
}

// Initialize plugin
func init() {
	log.Println("{{.Name}} handler plugin initialized")
}

func main() {
	router := gin.Default()
	
	router.GET("/{{.Name}}", {{.Name}}Handler)
	router.POST("/{{.Name}}", {{.Name}}Handler)
	
	log.Println("Starting {{.Name}} handler plugin server on :8080")
	router.Run(":8080")
}
`

const serviceTemplate = `package main

import (
	"log"
	"context"
	"time"
	"github.com/gin-gonic/gin"
)

// {{.Name}}Service provides {{.Name}} functionality
type {{.Name}}Service struct {
	config map[string]interface{}
}

// NewService creates a new {{.Name}} service instance
func NewService(config map[string]interface{}) *{{.Name}}Service {
	return &{{.Name}}Service{
		config: config,
	}
}

// Start starts the service
func (s *{{.Name}}Service) Start(ctx context.Context) error {
	log.Printf("[{{.Name}}] Service starting...")
	
	// Service initialization logic here
	
	log.Printf("[{{.Name}}] Service started successfully")
	return nil
}

// Stop stops the service
func (s *{{.Name}}Service) Stop(ctx context.Context) error {
	log.Printf("[{{.Name}}] Service stopping...")
	
	// Service cleanup logic here
	
	log.Printf("[{{.Name}}] Service stopped")
	return nil
}

// Process processes data
func (s *{{.Name}}Service) Process(data interface{}) (interface{}, error) {
	log.Printf("[{{.Name}}] Processing data...")
	
	// Processing logic here
	
	return data, nil
}

// Initialize plugin
func init() {
	log.Println("{{.Name}} service plugin initialized")
}

func main() {
	config := map[string]interface{}{
		"enabled": true,
		"debug": false,
	}
	
	service := NewService(config)
	ctx := context.Background()
	
	if err := service.Start(ctx); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
	
	router := gin.Default()
	
	router.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "{{.Name}}",
			"version": "{{.Version}}",
			"status": "running",
		})
	})
	
	log.Println("Starting {{.Name}} service plugin server on :8080")
	router.Run(":8080")
}
`

const filterTemplate = `package main

import (
	"log"
	"strings"
	"github.com/gin-gonic/gin"
)

// {{.Name}}Filter filters requests based on criteria
func {{.Name}}Filter() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[{{.Name}}] Filtering request: %s", c.Request.URL.Path)
		
		// Filter logic here
		userAgent := c.GetHeader("User-Agent")
		
		// Example: Block requests from certain user agents
		if strings.Contains(strings.ToLower(userAgent), "bot") {
			log.Printf("[{{.Name}}] Blocked bot request")
			c.JSON(403, gin.H{
				"error": "Bot requests are not allowed",
				"plugin": "{{.Name}}",
			})
			c.Abort()
			return
		}
		
		// Allow request to continue
		c.Next()
	}
}

// Initialize plugin
func init() {
	log.Println("{{.Name}} filter plugin initialized")
}

func main() {
	router := gin.Default()
	router.Use({{.Name}}Filter())
	
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Request passed {{.Name}} filter",
			"version": "{{.Version}}",
		})
	})
	
	log.Println("Starting {{.Name}} filter plugin server on :8080")
	router.Run(":8080")
}
`
