#!/bin/bash

# Git Log Analyzer Demo Script
# This script demonstrates various features of the git log analyzer

echo "=== Git Log Analyzer Demo ==="
echo ""

# Check if binary exists
if [ ! -f "./git-log-analyzer" ]; then
    echo "Building git-log-analyzer..."
    go build -o git-log-analyzer
fi

echo "1. Basic Analysis (current repository):"
echo "----------------------------------------"
./git-log-analyzer --web=false
echo ""

echo "2. Generate Web Report:"
echo "----------------------"
./git-log-analyzer --output-dir ./demo-reports --web
echo "Web report generated in ./demo-reports/"
echo ""

echo "3. Analysis with text output to file:"
echo "------------------------------------"
./git-log-analyzer --output demo-text-report.txt --web=false
echo "Text report saved. Here's a preview:"
head -15 demo-text-report.txt
echo ""

echo "4. Environment Variables Setup:"
echo "------------------------------"
echo "Copy env.sample to .env and configure:"
echo "cp env.sample .env"
echo ""
echo "Key environment variables:"
echo "- AI_API_KEY: Your OpenAI API key"
echo "- AI_API_ENDPOINT: Custom API endpoint"
echo "- AI_MODEL: Model to use (gpt-3.5-turbo, gpt-4, etc.)"
echo "- REPORT_OUTPUT_DIR: Default output directory"
echo ""

echo "5. AI Analysis (requires API key):"
echo "----------------------------------"
echo "To enable AI analysis:"
echo "export AI_API_KEY='your-openai-api-key'"
echo "export AI_MODEL='gpt-3.5-turbo'"
echo "Then run: ./git-log-analyzer --ai --web"
echo ""

echo "6. Advanced Usage Examples:"
echo "--------------------------"
echo "# Generate both text and web reports with AI:"
echo "./git-log-analyzer --ai --output full-report.txt --output-dir reports"
echo ""
echo "# Analyze different repo with custom output:"
echo "./git-log-analyzer --repo ../other-project --output-dir ../analysis"
echo ""
echo "# Generate web report and open in browser:"
echo "./git-log-analyzer --web --open"
echo ""

echo "7. Help and Configuration:"
echo "-------------------------"
./git-log-analyzer --help
echo ""

echo "8. File Structure:"
echo "-----------------"
echo "Generated reports include:"
echo "- index.html: Main web report"
echo "- styles.css: Styling"
echo "- charts.js: Interactive charts"
echo ""

echo "Demo completed! Check out the generated reports:"
echo "- ./demo-reports/ (Web report)"
echo "- demo-text-report.txt (Text report)"
echo ""
echo "To view the web report, open ./demo-reports/index.html in your browser"
