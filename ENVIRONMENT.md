# 环境变量配置指南

Git Log Analyzer 支持通过环境变量或 `.env` 文件进行配置。

## 环境变量加载机制

程序会按以下顺序查找并加载 `.env` 文件：

1. 当前工作目录的 `.env` 文件
2. 父目录的 `.env` 文件（适用于从子目录运行测试脚本的情况）
3. 祖父目录的 `.env` 文件

这样设计确保了无论从哪个目录运行程序，都能正确加载配置。

## 配置方式

### 方式1：直接设置环境变量
```bash
export AI_API_KEY="your-api-key"
export AI_MODEL="gpt-3.5-turbo"
./git-log-analyzer --repo . --ai
```

### 方式2：使用 .env 文件（推荐）
```bash
# 创建 .env 文件
cp .env.sample .env

# 编辑 .env 文件，设置你的API密钥
# AI_API_KEY=your-actual-api-key

# 运行程序
./git-log-analyzer --repo . --ai
```

## 支持的环境变量

- `AI_API_KEY`: AI API 密钥（必需，用于启用AI功能）
- `AI_API_ENDPOINT`: AI API 端点（可选，默认OpenAI）
- `AI_MODEL`: AI 模型（可选，默认 gpt-3.5-turbo）
- `AI_MAX_TOKENS`: 最大令牌数（可选，默认 2000）
- `AI_TEMPERATURE`: 温度参数（可选，默认 0.7）

## 优先级

环境变量的优先级为：
1. 系统环境变量（最高优先级）
2. .env 文件中的配置

如果系统中已经设置了某个环境变量，.env 文件中的同名变量将被忽略。

## 测试验证

你可以使用测试脚本验证配置是否生效：

```bash
# 测试AI功能（会自动加载.env文件）
./test/test-ai.sh

# 或者直接运行程序
./git-log-analyzer --repo . --ai
```

如果配置正确，程序将能够调用AI API进行分析。如果API密钥无效，会显示API错误信息而不是"环境变量未设置"的错误。
