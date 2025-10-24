package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var (
		testType = flag.String("type", "all", "Test type: unit, integration, all")
		verbose  = flag.Bool("v", false, "Verbose output")
		coverage = flag.Bool("cover", false, "Generate coverage report")
		package_ = flag.String("pkg", "", "Specific package to test")
	)
	flag.Parse()

	fmt.Println("Running tests...")
	fmt.Printf("Type: %s, Verbose: %v, Coverage: %v\n", *testType, *verbose, *coverage)

	var testDirs []string
	switch *testType {
	case "unit":
		testDirs = []string{"unit"}
	case "integration":
		testDirs = []string{"integration"}
	case "all":
		testDirs = []string{"unit", "integration"}
	default:
		fmt.Printf("Unknown test type: %s\n", *testType)
		os.Exit(1)
	}

	for _, dir := range testDirs {
		if err := runTests(dir, *verbose, *coverage, *package_); err != nil {
			fmt.Printf("Tests failed in %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	fmt.Println("All tests passed!")
}

func runTests(testDir string, verbose, coverage bool, pkg string) error {
	fmt.Printf("\n=== Running %s tests ===\n", testDir)
	
	testPath := filepath.Join("tests", testDir)
	if pkg != "" {
		testPath = filepath.Join(testPath, pkg)
	}

	args := []string{"test"}
	
	if verbose {
		args = append(args, "-v")
	}
	
	if coverage {
		coverProfile := fmt.Sprintf("coverage_%s.out", testDir)
		args = append(args, "-coverprofile="+coverProfile)
	}
	
	args = append(args, "./"+testPath+"/...")

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Printf("Running: go %s\n", strings.Join(args, " "))
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test command failed: %w", err)
	}

	if coverage {
		coverProfile := fmt.Sprintf("coverage_%s.out", testDir)
		if err := generateCoverageReport(coverProfile, testDir); err != nil {
			fmt.Printf("Warning: Failed to generate coverage report: %v\n", err)
		}
	}

	return nil
}

func generateCoverageReport(profileFile, testDir string) error {
	// Generate HTML coverage report
	htmlFile := fmt.Sprintf("coverage_%s.html", testDir)
	cmd := exec.Command("go", "tool", "cover", "-html="+profileFile, "-o", htmlFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate HTML coverage: %w", err)
	}
	
	fmt.Printf("Coverage report generated: %s\n", htmlFile)
	
	// Show coverage percentage
	cmd = exec.Command("go", "tool", "cover", "-func="+profileFile)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get coverage percentage: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			fmt.Printf("Total coverage: %s\n", strings.TrimSpace(line))
			break
		}
	}
	
	return nil
}