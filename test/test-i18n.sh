#!/bin/bash

# Git Log Analyzer - 国际化测试脚本
# 测试中英文报告生成功能

echo "=== Git Log Analyzer 国际化测试 ==="
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

# 清理之前的测试结果
echo "清理之前的测试结果..."
rm -rf i18n-test-zh/ i18n-test-en/
rm -f i18n-zh-*.txt i18n-en-*.txt
echo ""

echo "1. 测试中文报告生成:"
echo "----------------------------------------"

# 测试中文文本报告
echo "🔍 测试中文文本报告..."
"$BINARY_PATH" --repo "$PROJECT_DIR" --lang zh --web=false --output i18n-zh-text.txt
if [ $? -eq 0 ]; then
    echo "✅ 中文文本报告生成成功"
    
    # 检查中文内容
    if grep -q "Git 仓库分析报告" i18n-zh-text.txt; then
        echo "✅ 中文标题正确"
    else
        echo "❌ 中文标题错误"
    fi
    
    if grep -q "主要贡献者" i18n-zh-text.txt; then
        echo "✅ 中文章节标题正确"
    else
        echo "❌ 中文章节标题错误"
    fi
    
    if grep -q "次提交" i18n-zh-text.txt; then
        echo "✅ 中文单位正确"
    else
        echo "❌ 中文单位错误"
    fi
else
    echo "❌ 中文文本报告生成失败"
fi

# 测试中文网页报告
echo "🔍 测试中文网页报告..."
"$BINARY_PATH" --repo "$PROJECT_DIR" --lang zh --output-dir ./i18n-test-zh
if [ $? -eq 0 ]; then
    echo "✅ 中文网页报告生成成功"
    
    # 检查HTML中的中文内容
    if grep -q "Git 仓库分析报告" i18n-test-zh/index.html; then
        echo "✅ HTML中文标题正确"
    else
        echo "❌ HTML中文标题错误"
    fi
    
    if grep -q "主要贡献者" i18n-test-zh/index.html; then
        echo "✅ HTML中文章节正确"
    else
        echo "❌ HTML中文章节错误"
    fi
    
    if grep -q "lang=\"zh-CN\"" i18n-test-zh/index.html; then
        echo "✅ HTML语言属性正确"
    else
        echo "❌ HTML语言属性错误"
    fi
else
    echo "❌ 中文网页报告生成失败"
fi

echo ""
echo "2. 测试英文报告生成:"
echo "----------------------------------------"

# 测试英文文本报告
echo "🔍 测试英文文本报告..."
"$BINARY_PATH" --repo "$PROJECT_DIR" --lang en --web=false --output i18n-en-text.txt
if [ $? -eq 0 ]; then
    echo "✅ 英文文本报告生成成功"
    
    # 检查英文内容
    if grep -q "Git Repository Analysis Report" i18n-en-text.txt; then
        echo "✅ 英文标题正确"
    else
        echo "❌ 英文标题错误"
    fi
    
    if grep -q "Top Contributors" i18n-en-text.txt; then
        echo "✅ 英文章节标题正确"
    else
        echo "❌ 英文章节标题错误"
    fi
    
    if grep -q "commits" i18n-en-text.txt; then
        echo "✅ 英文单位正确"
    else
        echo "❌ 英文单位错误"
    fi
else
    echo "❌ 英文文本报告生成失败"
fi

# 测试英文网页报告
echo "🔍 测试英文网页报告..."
"$BINARY_PATH" --repo "$PROJECT_DIR" --lang en --output-dir ./i18n-test-en
if [ $? -eq 0 ]; then
    echo "✅ 英文网页报告生成成功"
    
    # 检查HTML中的英文内容
    if grep -q "Git Repository Analysis Report" i18n-test-en/index.html; then
        echo "✅ HTML英文标题正确"
    else
        echo "❌ HTML英文标题错误"
    fi
    
    if grep -q "Top Contributors" i18n-test-en/index.html; then
        echo "✅ HTML英文章节正确"
    else
        echo "❌ HTML英文章节错误"
    fi
    
    if grep -q "lang=\"en\"" i18n-test-en/index.html; then
        echo "✅ HTML语言属性正确"
    else
        echo "❌ HTML语言属性错误"
    fi
else
    echo "❌ 英文网页报告生成失败"
fi

echo ""
echo "3. 环境变量测试:"
echo "----------------------------------------"

# 测试通过环境变量设置语言
echo "🔍 测试环境变量语言设置..."
REPORT_LANGUAGE=en "$BINARY_PATH" --repo "$PROJECT_DIR" --web=false --output i18n-env-test.txt
if [ $? -eq 0 ]; then
    if grep -q "Git Repository Analysis Report" i18n-env-test.txt; then
        echo "✅ 环境变量语言设置正确"
    else
        echo "❌ 环境变量语言设置错误"
    fi
else
    echo "❌ 环境变量测试失败"
fi

echo ""
echo "4. 对比测试:"
echo "----------------------------------------"

echo "生成的测试文件:"
echo "📄 中文文本报告: i18n-zh-text.txt"
echo "📄 英文文本报告: i18n-en-text.txt"
echo "🌐 中文网页报告: i18n-test-zh/"
echo "🌐 英文网页报告: i18n-test-en/"

echo ""
echo "内容对比示例:"
echo "中文: $(grep "=== " i18n-zh-text.txt | head -1)"
echo "英文: $(grep "=== " i18n-en-text.txt | head -1)"

echo ""
echo "国际化测试完成！🎉"
echo ""
echo "要查看网页报告，请运行："
echo "open i18n-test-zh/index.html  # 中文版本"
echo "open i18n-test-en/index.html  # 英文版本"
