package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git-log-analyzer/internal/ai"
	"git-log-analyzer/internal/analyzer"
	"git-log-analyzer/internal/git"
	"git-log-analyzer/internal/report"
)

var cfgFile string
var repoPath string
var useAI bool
var outputFile string
var generateWeb bool
var outputDir string
var openBrowser bool
var reportLanguage string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-log-analyzer",
	Short: "A Git log analysis tool",
	Long: `Git Log Analyzer is a command-line tool that analyzes Git commit logs
and provides insights about your repository's development history.

This tool can:
- Analyze commits by author, time, and other dimensions
- Generate statistical reports
- Use AI models for advanced analysis

Environment variables for AI analysis:
- AI_API_ENDPOINT: API endpoint (default: https://api.openai.com/v1/chat/completions)
- AI_API_KEY: API key (required for AI analysis)
- AI_MODEL: Model to use (default: gpt-3.5-turbo)
- AI_MAX_TOKENS: Maximum tokens (default: 2000)
- AI_TEMPERATURE: Temperature setting (default: 0.7)

Environment variables for report customization:
- REPORT_LANGUAGE: Report language (zh/en, default: zh)`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := analyzeGitLog(repoPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-log-analyzer.yaml)")
	rootCmd.PersistentFlags().StringVarP(&repoPath, "repo", "r", "./", "path to git repository")
	rootCmd.PersistentFlags().BoolVar(&useAI, "ai", false, "enable AI-powered analysis")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file for the text report")
	rootCmd.PersistentFlags().BoolVar(&generateWeb, "web", true, "generate web-based HTML report")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output-dir", getEnv("REPORT_OUTPUT_DIR", "./analysis-reports"), "output directory for reports")
	rootCmd.PersistentFlags().BoolVar(&openBrowser, "open", getEnvBool("AUTO_OPEN_BROWSER", false), "automatically open web report in browser")
	rootCmd.PersistentFlags().StringVarP(&reportLanguage, "lang", "l", getEnv("REPORT_LANGUAGE", "zh"), "report language (zh/en)")

	// Bind flags to viper
	viper.BindPFlag("repo", rootCmd.PersistentFlags().Lookup("repo"))
	viper.BindPFlag("ai", rootCmd.PersistentFlags().Lookup("ai"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("web", rootCmd.PersistentFlags().Lookup("web"))
	viper.BindPFlag("output-dir", rootCmd.PersistentFlags().Lookup("output-dir"))
	viper.BindPFlag("open", rootCmd.PersistentFlags().Lookup("open"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".git-log-analyzer")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func analyzeGitLog(repoPath string) error {
	// Set language from command line flag
	if reportLanguage != "" {
		os.Setenv("REPORT_LANGUAGE", reportLanguage)
	}
	
	fmt.Printf("Analyzing git repository at: %s\n", repoPath)
	
	// Verify git installation
	if !git.IsGitInstalled() {
		return fmt.Errorf("git is not installed or not available in PATH")
	}
	
	// Create analyzer
	a := analyzer.NewAnalyzer(repoPath)
	
	// Perform analysis
	fmt.Println("Performing git log analysis...")
	stats, err := a.Analyze()
	if err != nil {
		return fmt.Errorf("failed to analyze repository: %v", err)
	}
	
	// Generate basic report
	basicReport := stats.GenerateReport()
	
	var finalReport string
	var aiAnalysis string
	
	// AI analysis if enabled
	if useAI {
		fmt.Println("Performing AI-powered analysis...")
		aiClient, err := ai.NewAIClient()
		if err != nil {
			fmt.Printf("Warning: AI analysis failed: %v\n", err)
			fmt.Println("Continuing with basic analysis only...")
			finalReport = basicReport
		} else {
			aiResult, err := aiClient.AnalyzeWithAI(stats, basicReport)
			if err != nil {
				fmt.Printf("Warning: AI analysis failed: %v\n", err)
				fmt.Println("Continuing with basic analysis only...")
				finalReport = basicReport
			} else {
				aiAnalysis = aiResult
				finalReport = basicReport + "\n\n=== AI-Powered Analysis ===\n" + aiAnalysis
			}
		}
	} else {
		finalReport = basicReport
	}
	
	// Generate web report
	if generateWeb {
		fmt.Println("Generating web report...")
		webGen := report.NewWebReportGenerator(outputDir)
		projectName := filepath.Base(repoPath)
		if projectName == "." || projectName == "" {
			projectName = "Current Repository"
		}
		
		err := webGen.GenerateReport(stats, aiAnalysis, projectName)
		if err != nil {
			fmt.Printf("Warning: Failed to generate web report: %v\n", err)
		} else {
			reportPath := webGen.GetReportPath()
			fmt.Printf("Web report generated: %s\n", reportPath)
			
			if openBrowser {
				openWebReport(reportPath)
			}
		}
	}
	
	// Output text results
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(finalReport), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}
		fmt.Printf("Text report saved to: %s\n", outputFile)
	} else {
		fmt.Println(finalReport)
	}
	
	return nil
}

// openWebReport opens the web report in the default browser
func openWebReport(reportPath string) {
	absPath, err := filepath.Abs(reportPath)
	if err != nil {
		fmt.Printf("Warning: Could not get absolute path for report: %v\n", err)
		return
	}
	
	fmt.Printf("Opening web report in browser: file://%s\n", absPath)
	// Note: This is a simplified version. In a real implementation,
	// you might want to use a library like "github.com/pkg/browser"
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets environment variable as boolean with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
