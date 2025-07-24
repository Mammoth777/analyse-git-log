#!/bin/bash

# Git Log Analyzer - AI功能测试脚本
# 测试AI分析功能是否正常工作

echo "=== Git Log Analyzer AI 测试 ==="
echo ""

# 检查是否存在主程序
if [ ! -f "../git-log-analyzer" ]; then
    echo "构建主程序..."
    cd .. && go build -o git-log-analyzer && cd test
fi

# 检查环境变量
echo "1. 检查AI配置:"
echo "----------------------------------------"
if [ -z "$AI_API_KEY" ]; then
    echo "❌ AI_API_KEY 未设置"
    echo "请设置环境变量: export AI_API_KEY='your-api-key'"
    echo ""
    echo "或者从 .env 文件加载:"
    echo "export \$(cat ../.env | xargs) 2>/dev/null"
    echo ""
else
    echo "✅ AI_API_KEY 已设置: ${AI_API_KEY:0:8}..."
fi

echo "AI_API_ENDPOINT: ${AI_API_ENDPOINT:-默认OpenAI}"
echo "AI_MODEL: ${AI_MODEL:-gpt-3.5-turbo}"
echo "AI_MAX_TOKENS: ${AI_MAX_TOKENS:-2000}"
echo "AI_TEMPERATURE: ${AI_TEMPERATURE:-0.7}"
echo ""

# 测试基础分析
echo "2. 基础分析测试:"
echo "----------------------------------------"
../git-log-analyzer --repo .. --web=false --output basic-test.txt
if [ $? -eq 0 ]; then
    echo "✅ 基础分析成功"
else
    echo "❌ 基础分析失败"
fi
echo ""

# 测试AI分析（如果配置了API密钥）
if [ ! -z "$AI_API_KEY" ]; then
    echo "3. AI分析测试:"
    echo "----------------------------------------"
    ../git-log-analyzer --repo .. --ai --web=false --output ai-test.txt
    if [ $? -eq 0 ]; then
        echo "✅ AI分析成功"
        echo "检查AI分析结果:"
        if grep -q "AI-Powered Analysis" ai-test.txt; then
            echo "✅ AI分析内容已生成"
        else
            echo "⚠️  AI分析内容可能为空"
        fi
    else
        echo "❌ AI分析失败"
    fi
else
    echo "3. AI分析测试: 跳过（未配置API密钥）"
fi
echo ""

# 测试网页报告生成
echo "4. 网页报告测试:"
echo "----------------------------------------"
../git-log-analyzer --repo .. --output-dir ./web-test-reports
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
if [ ! -z "$AI_API_KEY" ]; then
    echo "- ai-test.txt (AI分析)"
fi
echo "- web-test-reports/ (网页报告)"
