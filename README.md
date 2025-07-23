# Git Log Analyzer

一个基于Golang的git log分析命令行工具，能够分析Git仓库的提交历史并提供详细的统计信息和AI驱动的分析。

## 功能特性

1. **Help功能** - 完整的命令行帮助和使用说明
2. **仓库路径参数** - 支持指定Git仓库路径，默认为当前目录（"./"）
3. **Git验证** - 自动验证Git是否安装，并通过Git命令获取日志
4. **多维度统计** - 根据人员、时间等多个维度统计数据进行初步分析
5. **AI分析** - 从环境变量获取大模型配置，调用AI获取进一步分析

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

#### 基本使用

```bash
# 分析当前目录的Git仓库
./git-log-analyzer

# 分析指定路径的Git仓库
./git-log-analyzer --repo /path/to/git/repository

# 显示帮助信息
./git-log-analyzer --help
```

#### 高级功能

```bash
# 启用AI分析（需要配置环境变量）
./git-log-analyzer --ai

# 将报告输出到文件
./git-log-analyzer --output report.txt

# 组合使用
./git-log-analyzer --repo /path/to/repo --ai --output analysis.txt
```

### AI分析配置

要使用AI分析功能，需要设置以下环境变量：

```bash
export AI_API_KEY="your-api-key"                                    # 必需
export AI_API_ENDPOINT="https://api.openai.com/v1/chat/completions" # 可选，默认OpenAI
export AI_MODEL="gpt-3.5-turbo"                                     # 可选，默认值
export AI_MAX_TOKENS="2000"                                         # 可选，默认值
export AI_TEMPERATURE="0.7"                                         # 可选，默认值
```

### 示例

```bash
# 设置API密钥
export AI_API_KEY="sk-your-openai-api-key"

# 运行完整分析
./git-log-analyzer --repo ~/my-project --ai --output detailed-report.txt
```

## 输出报告

工具会生成包含以下内容的分析报告：

### 基础统计
- 总提交数
- 活跃时间段
- 活跃天数、周数、月数

### 贡献者分析
- 按提交数排序的贡献者
- 每个贡献者的代码行数变化（增加/删除）

### 时间模式分析
- 最活跃的提交时间
- 一周中各天的提交分布

### 文件修改统计
- 最常修改的文件
- 文件修改频率

### AI深度分析（可选）
- 开发模式分析
- 团队协作洞察
- 代码质量观察
- 生产力趋势
- 改进建议

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
