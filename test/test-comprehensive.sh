#!/bin/bash

# Git Log Analyzer - 综合测试脚本
# 测试所有主要功能

echo "=== Git Log Analyzer 综合测试 ==="
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
rm -rf comprehensive-test-reports/
rm -f comprehensive-test.txt
echo ""

echo "1. 基础功能测试:"
echo "----------------------------------------"

# 测试基础分析
echo "🔍 测试基础分析..."
"$BINARY_PATH" --repo "$PROJECT_DIR" --web=false --output comprehensive-test.txt
if [ $? -eq 0 ]; then
    echo "✅ 基础分析成功"
else
    echo "❌ 基础分析失败"
    exit 1
fi

# 测试网页报告生成
echo "🔍 测试网页报告生成..."
"$BINARY_PATH" --repo "$PROJECT_DIR" --output-dir comprehensive-test-reports --web
if [ $? -eq 0 ]; then
    echo "✅ 网页报告生成成功"
else
    echo "❌ 网页报告生成失败"
    exit 1
fi

echo ""
echo "2. 数据完整性检查:"
echo "----------------------------------------"

# 检查基础数据
if [ -f "comprehensive-test.txt" ]; then
    commits=$(grep "Total Commits:" comprehensive-test.txt | sed 's/.*: //')
    authors=$(grep -c "commits (" comprehensive-test.txt)
    
    echo "📊 总提交数: $commits"
    echo "👥 贡献者数: $authors"
    
    if [ "$commits" -gt 0 ]; then
        echo "✅ 提交数据正常"
    else
        echo "❌ 提交数据异常"
    fi
    
    if [ "$authors" -gt 0 ]; then
        echo "✅ 贡献者数据正常"
    else
        echo "❌ 贡献者数据异常"
    fi
else
    echo "❌ 文本报告文件不存在"
fi

# 检查网页报告文件
echo ""
echo "3. 网页报告文件检查:"
echo "----------------------------------------"

required_files=("index.html" "styles.css" "charts.js")
for file in "${required_files[@]}"; do
    if [ -f "comprehensive-test-reports/$file" ]; then
        size=$(ls -lh "comprehensive-test-reports/$file" | awk '{print $5}')
        echo "✅ $file 存在 (大小: $size)"
    else
        echo "❌ $file 缺失"
    fi
done

# 检查HTML中的数据
if [ -f "comprehensive-test-reports/index.html" ]; then
    echo ""
    echo "4. HTML数据验证:"
    echo "----------------------------------------"
    
    if grep -q "reportData" comprehensive-test-reports/index.html; then
        echo "✅ JavaScript数据对象存在"
        
        # 提取并验证JSON数据
        if grep -q '"Name":' comprehensive-test-reports/index.html; then
            echo "✅ 作者数据格式正确"
        else
            echo "❌ 作者数据格式错误"
        fi
        
        if grep -q '"Hour":' comprehensive-test-reports/index.html; then
            echo "✅ 小时数据格式正确"
        else
            echo "❌ 小时数据格式错误"
        fi
        
        if grep -q '"Day":' comprehensive-test-reports/index.html; then
            echo "✅ 日期数据格式正确"
        else
            echo "❌ 日期数据格式错误"
        fi
        
        if grep -q '"Date":' comprehensive-test-reports/index.html; then
            echo "✅ 时间线数据格式正确"
        else
            echo "❌ 时间线数据格式错误"
        fi
    else
        echo "❌ JavaScript数据对象缺失"
    fi
fi

echo ""
echo "5. 图表元素检查:"
echo "----------------------------------------"

chart_elements=("authorsChart" "timelineChart" "hourlyChart" "dailyChart")
for chart in "${chart_elements[@]}"; do
    if grep -q "id=\"$chart\"" comprehensive-test-reports/index.html; then
        echo "✅ $chart 元素存在"
    else
        echo "❌ $chart 元素缺失"
    fi
done

echo ""
echo "6. CSS和JS功能检查:"
echo "----------------------------------------"

# 检查CSS样式
if grep -q ".chart-container" comprehensive-test-reports/styles.css; then
    echo "✅ 图表容器样式存在"
else
    echo "❌ 图表容器样式缺失"
fi

# 检查JavaScript函数
if grep -q "function initCharts" comprehensive-test-reports/charts.js; then
    echo "✅ 图表初始化函数存在"
else
    echo "❌ 图表初始化函数缺失"
fi

if grep -q "console.log" comprehensive-test-reports/charts.js; then
    echo "✅ 调试日志存在"
else
    echo "❌ 调试日志缺失"
fi

echo ""
echo "7. 最终报告:"
echo "----------------------------------------"

echo "生成的文件:"
echo "📄 文本报告: comprehensive-test.txt"
echo "🌐 网页报告: comprehensive-test-reports/"
echo "   ├── index.html"
echo "   ├── styles.css"
echo "   └── charts.js"

echo ""
echo "要查看网页报告，请在浏览器中打开："
echo "file://$(pwd)/comprehensive-test-reports/index.html"

echo ""
echo "或运行以下命令自动打开："
echo "open comprehensive-test-reports/index.html"

echo ""
echo "测试完成！🎉"
