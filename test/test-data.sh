#!/bin/bash

# Git Log Analyzer - 数据验证测试脚本
# 检查生成的报告中数据显示是否正常

echo "=== Git Log Analyzer 数据验证测试 ==="
echo ""

# 检查是否存在主程序
if [ ! -f "../git-log-analyzer" ]; then
    echo "构建主程序..."
    cd .. && go build -o git-log-analyzer && cd test
fi

echo "1. 生成测试报告:"
echo "----------------------------------------"
../git-log-analyzer --repo .. --output-dir ./data-test-reports --output data-test.txt

echo ""
echo "2. 检查文本报告数据:"
echo "----------------------------------------"
if [ -f "data-test.txt" ]; then
    echo "✅ 文本报告已生成"
    
    # 检查各项数据
    if grep -q "Total Commits:" data-test.txt; then
        commits=$(grep "Total Commits:" data-test.txt | sed 's/.*: //')
        echo "✅ 总提交数: $commits"
    else
        echo "❌ 总提交数缺失"
    fi
    
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
    echo "❌ 文本报告生成失败"
fi

echo ""
echo "3. 检查网页报告数据:"
echo "----------------------------------------"
if [ -f "./data-test-reports/index.html" ]; then
    echo "✅ 网页报告已生成"
    
    # 检查HTML中是否包含图表数据
    if grep -q "authorsChart" ./data-test-reports/index.html; then
        echo "✅ 作者图表代码存在"
    else
        echo "❌ 作者图表代码缺失"
    fi
    
    if grep -q "timelineChart" ./data-test-reports/index.html; then
        echo "✅ 时间线图表代码存在"
    else
        echo "❌ 时间线图表代码缺失"
    fi
    
    if grep -q "hourlyChart" ./data-test-reports/index.html; then
        echo "✅ 小时图表代码存在"
    else
        echo "❌ 小时图表代码缺失"
    fi
    
    if grep -q "dailyChart" ./data-test-reports/index.html; then
        echo "✅ 日常图表代码存在"
    else
        echo "❌ 日常图表代码缺失"
    fi
    
    # 检查JavaScript数据
    if grep -q "reportData" ./data-test-reports/index.html; then
        echo "✅ JavaScript报告数据存在"
        
        # 提取并显示数据结构
        echo ""
        echo "数据结构检查:"
        grep -A 10 "const reportData" ./data-test-reports/index.html
    else
        echo "❌ JavaScript报告数据缺失"
    fi
    
else
    echo "❌ 网页报告生成失败"
fi

echo ""
echo "4. 调试信息:"
echo "----------------------------------------"
echo "当前Git仓库信息:"
git -C .. log --oneline -5
echo ""
echo "Git仓库统计:"
echo "总提交数: $(git -C .. rev-list --all --count)"
echo "作者数量: $(git -C .. log --format='%an' | sort -u | wc -l)"
echo "活跃天数: $(git -C .. log --format='%ai' | cut -d' ' -f1 | sort -u | wc -l)"

echo ""
echo "测试完成！请查看生成的文件以进一步分析问题。"
