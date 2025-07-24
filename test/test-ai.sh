#!/bin/bash

# Git Log Analyzer - AI功能测试脚本  
# 测试AI分析功能是否正常工作
# 注意：程序本身负责环境变量验证，测试脚本只验证功能

echo "=== Git Log Analyzer AI 测试 ==="
echo ""

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY_PATH="$PROJECT_DIR/git-log-analyzer"

# 切换到项目根目录
cd "$PROJECT_DIR"

# 检查是否存在主程序
if [ ! -f "$BINARY_PATH" ]; then
    echo "构建主程序..."
    go build -o git-log-analyzer
    if [ $? -ne 0 ]; then
        echo "❌ 构建失败"
        exit 1
    fi
    echo "✅ 构建成功"
fi

# 切换回test目录
cd "$SCRIPT_DIR"

# 测试基础分析
echo "1. 基础分析测试:"
echo "----------------------------------------"
"$BINARY_PATH" --repo "$PROJECT_DIR" --web=false --output basic-test.txt
if [ $? -eq 0 ]; then
    echo "✅ 基础分析成功"
else
    echo "❌ 基础分析失败"
fi
echo ""

# 测试AI分析
echo "2. AI分析测试:"
echo "----------------------------------------"
"$BINARY_PATH" --repo "$PROJECT_DIR" --ai --web=false --output ai-test.txt
if [ $? -eq 0 ]; then
    echo "✅ AI分析成功"
    echo "检查AI分析结果:"
    if grep -q "AI-Powered Analysis" ai-test.txt; then
        echo "✅ AI分析内容已生成"
    else
        echo "⚠️  AI分析内容可能为空"
    fi
else
    echo "❌ AI分析失败（可能需要配置AI_API_KEY）"
    echo "如需使用AI功能，请："
    echo "1. 设置环境变量: export AI_API_KEY='your-api-key'"
    echo "2. 或创建 .env 文件: AI_API_KEY=your-api-key"
fi
echo ""

# 测试网页报告生成
echo "3. 网页报告测试:"
echo "----------------------------------------"
"$BINARY_PATH" --repo "$PROJECT_DIR" --output-dir ./web-test-reports
if [ $? -eq 0 ]; then
    echo "✅ 网页报告生成成功"
    if [ -f "./web-test-reports/index.html" ]; then
        echo "✅ HTML文件已生成"
    else
        echo "❌ HTML文件生成失败"
    fi
else
    echo "❌ 网页报告生成失败"
fi
echo ""

echo "测试完成！"
echo "生成的测试文件:"
echo "- basic-test.txt (基础分析)"
echo "- ai-test.txt (AI分析)"
echo "- web-test-reports/ (网页报告)"
