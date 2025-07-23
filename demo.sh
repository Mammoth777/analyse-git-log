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
./git-log-analyzer
echo ""

echo "2. Analysis with output to file:"
echo "--------------------------------"
./git-log-analyzer --output demo-report.txt
echo "Report saved. Here's a preview:"
head -15 demo-report.txt
echo ""

echo "3. Analysis of different repository path:"
echo "----------------------------------------"
echo "Usage: ./git-log-analyzer --repo /path/to/other/repo"
echo ""

echo "4. AI Analysis (requires API key):"
echo "----------------------------------"
echo "To enable AI analysis, set environment variables:"
echo "export AI_API_KEY='your-openai-api-key'"
echo "export AI_MODEL='gpt-3.5-turbo'"
echo "Then run: ./git-log-analyzer --ai"
echo ""

echo "5. Help and Documentation:"
echo "-------------------------"
./git-log-analyzer --help
echo ""

echo "6. Advanced Usage Examples:"
echo "--------------------------"
echo "# Analyze with AI and save to file:"
echo "./git-log-analyzer --ai --output full-analysis.txt"
echo ""
echo "# Analyze different repo with AI:"
echo "./git-log-analyzer --repo ../other-project --ai"
echo ""

echo "Demo completed! Check out the generated reports:"
echo "- demo-report.txt"
echo "- analysis-report.txt (from previous run)"
