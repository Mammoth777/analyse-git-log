# Git Log Analyzer

一个基于Golang的git log分析命令行工具，能够分析Git仓库的提交历史并提供详细的统计信息和AI驱动的分析。

## 功能特性

1. **Help功能** - 完整的命令行帮助和使用说明
2. **仓库路径参数** - 支持指定Git仓库路径，默认为当前目录（"./"）
3. **Git验证** - 自动验证Git是否安装，并通过Git命令获取日志
4. **多维度统计** - 根据人员、时间等多个维度统计数据进行初步分析
5. **AI分析** - 从环境变量获取大模型配置，调用AI获取进一步分析
6. **网页报告** - 生成美观的HTML报告，包含交互式图表
7. **灵活配置** - 支持环境变量和命令行参数配置

## 安装和使用

### 前置要求

- Go 1.21+
- Git 命令行工具

### 安装依赖

```bash
go mod tidy
```

### 构建

```bash
go build -o git-log-analyzer
```

### 使用方法

#### 环境配置

首先复制环境变量示例文件并根据需要配置：

```bash
cp env.sample .env
# 编辑 .env 文件，设置API密钥等配置
```

#### 基本使用

```bash
# 分析当前目录的Git仓库（生成网页报告）
./git-log-analyzer

# 分析指定路径的Git仓库
./git-log-analyzer --repo /path/to/git/repository

# 只生成文本报告
./git-log-analyzer --web=false

# 显示帮助信息
./git-log-analyzer --help
```

#### 高级功能

```bash
# 启用AI分析（需要配置环境变量）
./git-log-analyzer --ai

# 自定义输出目录
./git-log-analyzer --output-dir ./my-reports

# 将文本报告输出到文件
./git-log-analyzer --output report.txt

# 生成网页报告并自动打开浏览器
./git-log-analyzer --web --open

# 组合使用
./git-log-analyzer --repo /path/to/repo --ai --output text-report.txt --output-dir web-reports
```

### AI分析配置

要使用AI分析功能，需要配置环境变量。复制 `env.sample` 为 `.env` 并设置：

```bash
# 必需 - API密钥
AI_API_KEY=sk-your-api-key

# 可选 - 自定义配置
AI_API_ENDPOINT=https://api.openai.com/v1/chat/completions
AI_MODEL=gpt-3.5-turbo
AI_MAX_TOKENS=2000
AI_TEMPERATURE=0.7

# 可选 - 报告配置
REPORT_OUTPUT_DIR=./analysis-reports
GENERATE_WEB_REPORT=true
AUTO_OPEN_BROWSER=false
```

### 输出报告

工具支持两种报告格式：

#### 1. 网页报告（默认）
- 生成美观的HTML报告，包含交互式图表
- 文件：`index.html`, `styles.css`, `charts.js`
- 支持响应式设计，适配移动设备

#### 2. 文本报告
- 传统的文本格式报告
- 适合命令行查看和自动化处理

### 示例

```bash
# 设置API密钥
export AI_API_KEY="sk-your-openai-api-key"

# 运行完整分析
./git-log-analyzer --repo ~/my-project --ai --output detailed-report.txt
```

## 输出报告

工具会生成包含以下内容的分析报告：

### 网页报告特性
- **交互式图表**: 使用Chart.js创建的动态图表
- **响应式设计**: 适配桌面和移动设备
- **现代UI**: 美观的卡片式布局和渐变设计
- **数据可视化**: 
  - 贡献者分布饼图
  - 提交时间线图
  - 小时活跃度柱状图
  - 每周活跃度极坐标图

### 报告内容
- **基础统计**: 总提交数、活跃时间段、活跃天数等
- **贡献者分析**: 按提交数排序，包含代码行数变化
- **时间模式分析**: 最活跃的提交时间和周期模式
- **文件修改统计**: 最常修改的文件和修改频率
- **AI深度分析**（可选）: 开发模式、团队协作、改进建议

## 项目结构

```
git-log-analyzer/
├── main.go                    # 程序入口
├── cmd/
│   └── root.go               # 命令行界面
├── internal/
│   ├── git/
│   │   └── git.go           # Git操作和日志解析
│   ├── analyzer/
│   │   └── analyzer.go      # 统计分析逻辑
│   └── ai/
│       └── ai.go            # AI分析集成
├── go.mod                   # Go模块定义
└── README.md               # 项目说明
```

## 开发

### 添加新功能

1. Git相关功能：修改 `internal/git/git.go`
2. 分析算法：修改 `internal/analyzer/analyzer.go`
3. AI集成：修改 `internal/ai/ai.go`
4. 命令行参数：修改 `cmd/root.go`

### 测试

```bash
# 运行测试
go test ./...

# 生成测试覆盖率报告
go test -cover ./...
```

## 许可证

本项目采用MIT许可证。

## 贡献

欢迎提交Issue和Pull Request来改进这个工具！
# Additional Documentation
