# Git Log Analyzer 测试套件

这个目录包含了Git Log Analyzer的各种测试脚本，用于验证工具的功能和数据输出。

## 测试脚本说明

### 1. `test-ai.sh` - AI功能测试
测试AI分析功能是否正常工作。

**功能:**
- 检查AI配置环境变量
- 测试基础分析功能
- 测试AI分析功能（如果配置了API密钥）
- 验证网页报告生成

**使用方法:**
```bash
# 设置环境变量（可选）
export AI_API_KEY="your-api-key"

# 运行测试
./test-ai.sh
```

### 2. `test-data.sh` - 数据验证测试
检查生成报告中的数据显示是否正常。

**功能:**
- 验证文本报告数据完整性
- 检查网页报告HTML结构
- 验证JavaScript数据格式
- 输出调试信息

**使用方法:**
```bash
./test-data.sh
```

### 3. `test-comprehensive.sh` - 综合测试
全面测试所有功能模块。

**功能:**
- 基础功能测试
- 数据完整性检查
- 网页报告文件验证
- HTML数据格式验证
- 图表元素检查
- CSS和JS功能检查

**使用方法:**
```bash
./test-comprehensive.sh
```

## 测试输出

所有测试脚本都会在当前目录下生成相应的测试文件：

### 文件结构
```
test/
├── README.md                    # 本文件
├── test-ai.sh                   # AI功能测试脚本
├── test-data.sh                 # 数据验证测试脚本
├── test-comprehensive.sh        # 综合测试脚本
├── basic-test.txt               # 基础分析结果
├── ai-test.txt                  # AI分析结果（如果启用）
├── data-test.txt                # 数据验证测试结果
├── comprehensive-test.txt       # 综合测试结果
├── web-test-reports/            # AI测试的网页报告
├── data-test-reports/           # 数据验证的网页报告
└── comprehensive-test-reports/  # 综合测试的网页报告
```

### 网页报告结构
```
*-reports/
├── index.html      # 主报告页面
├── styles.css      # 样式文件
└── charts.js       # 图表脚本
```

## 问题排查

如果测试失败，请检查以下项目：

### 1. 构建问题
```bash
# 确保主程序已构建
cd .. && go build -o git-log-analyzer
```

### 2. 权限问题
```bash
# 确保测试脚本有执行权限
chmod +x test-*.sh
```

### 3. 环境变量问题
```bash
# 检查AI配置（可选）
echo $AI_API_KEY
echo $AI_API_ENDPOINT
```

### 4. Git仓库问题
```bash
# 确保在Git仓库中运行测试
git status
```

## 测试数据问题排查

如果网页报告中的图表显示空白，可能的原因：

1. **数据格式问题**: 运行 `test-data.sh` 检查JSON数据格式
2. **JavaScript错误**: 在浏览器开发者工具中查看控制台错误
3. **Chart.js加载问题**: 检查网络连接，Chart.js从CDN加载
4. **数据为空**: 确保Git仓库有足够的提交历史

## 添加新测试

要添加新的测试脚本：

1. 创建新的shell脚本文件
2. 添加执行权限: `chmod +x new-test.sh`
3. 按照现有脚本的格式编写测试逻辑
4. 更新本README文件

## 最佳实践

- 在每次修改代码后运行 `test-comprehensive.sh`
- 定期清理测试生成的文件
- 在提交代码前确保所有测试通过
- 在不同环境下测试以确保兼容性
