package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	
	"git-log-analyzer/internal/analyzer"
	"git-log-analyzer/internal/ai"
	"git-log-analyzer/internal/git"
)

var cfgFile string
var repoPath string
var useAI bool
var outputFile string

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
- AI_TEMPERATURE: Temperature setting (default: 0.7)`,
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
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file for the report")

	// Bind flags to viper
	viper.BindPFlag("repo", rootCmd.PersistentFlags().Lookup("repo"))
	viper.BindPFlag("ai", rootCmd.PersistentFlags().Lookup("ai"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
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
	
	// AI analysis if enabled
	if useAI {
		fmt.Println("Performing AI-powered analysis...")
		aiClient, err := ai.NewAIClient()
		if err != nil {
			fmt.Printf("Warning: AI analysis failed: %v\n", err)
			fmt.Println("Continuing with basic analysis only...")
			finalReport = basicReport
		} else {
			aiAnalysis, err := aiClient.AnalyzeWithAI(stats, basicReport)
			if err != nil {
				fmt.Printf("Warning: AI analysis failed: %v\n", err)
				fmt.Println("Continuing with basic analysis only...")
				finalReport = basicReport
			} else {
				finalReport = basicReport + "\n\n=== AI-Powered Analysis ===\n" + aiAnalysis
			}
		}
	} else {
		finalReport = basicReport
	}
	
	// Output results
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(finalReport), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}
		fmt.Printf("Report saved to: %s\n", outputFile)
	} else {
		fmt.Println(finalReport)
	}
	
	return nil
}
