#!/bin/bash

# Git Log Analyzer - 数据验证测试脚本
# 检查生成的报告中数据显示是否正常

echo "=== Git Log Analyzer 数据验证测试 ==="
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

echo "1. 生成测试报告:"
echo "----------------------------------------"
"$BINARY_PATH" --repo "$PROJECT_DIR" --output-dir ./data-test-reports --output data-test.txt

echo ""
echo "2. 检查文本报告数据:"
echo "----------------------------------------"

if [ -f "data-test.txt" ]; then
    echo "✅ 文本报告已生成"
    
    # 检查基本数据
    commits=$(grep "Total Commits:" data-test.txt | awk '{print $3}')
    echo "✅ 总提交数: $commits"
    
    if grep -q "Top Contributors" data-test.txt; then
        echo "✅ 贡献者数据存在"
    else
        echo "❌ 贡献者数据缺失"
    fi
    
    if grep -q "Most Active Hours" data-test.txt; then
        echo "✅ 活跃小时数据存在"
    else
        echo "❌ 活跃小时数据缺失"
    fi
    
    if grep -q "Most Modified Files" data-test.txt; then
        echo "✅ 文件修改数据存在"
    else
        echo "❌ 文件修改数据缺失"
    fi
else
    echo "❌ 文本报告未生成"
fi

echo ""
echo "3. 检查网页报告数据:"
echo "----------------------------------------"

if [ -f "data-test-reports/index.html" ]; then
    echo "✅ 网页报告已生成"
    
    # 检查图表相关代码
    if grep -q "authorsChart" data-test-reports/index.html; then
        echo "✅ 作者图表代码存在"
    else
        echo "❌ 作者图表代码缺失"
    fi
    
    if grep -q "timelineChart" data-test-reports/index.html; then
        echo "✅ 时间线图表代码存在"
    else
        echo "❌ 时间线图表代码缺失"
    fi
    
    if grep -q "hourlyChart" data-test-reports/index.html; then
        echo "✅ 小时图表代码存在"
    else
        echo "❌ 小时图表代码缺失"
    fi
    
    if grep -q "dailyChart" data-test-reports/index.html; then
        echo "✅ 日常图表代码存在"
    else
        echo "❌ 日常图表代码缺失"
    fi
    
    # 检查JavaScript数据对象
    if grep -q "reportData" data-test-reports/index.html; then
        echo "✅ JavaScript报告数据存在"
        echo ""
        echo "数据结构检查:"
        grep -A 10 "const reportData" data-test-reports/index.html
    else
        echo "❌ JavaScript报告数据缺失"
    fi
else
    echo "❌ 网页报告未生成"
fi

echo ""
echo "4. 调试信息:"
echo "----------------------------------------"

echo "当前Git仓库信息:"
git log --oneline -5

echo ""
echo "Git仓库统计:"
echo "总提交数: $(git log --oneline | wc -l | tr -d ' ')"
echo "作者数量: $(git log --pretty=format:'%an' | sort -u | wc -l | tr -d ' ')"
echo "活跃天数: $(git log --pretty=format:'%ad' --date=short | sort -u | wc -l | tr -d ' ')"

echo ""
echo "测试完成！请查看生成的文件以进一步分析问题。"
