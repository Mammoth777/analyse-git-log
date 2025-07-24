# 程序分发说明

## 当前状态：模板已嵌入

经过优化，程序现在使用 Go 1.16+ 的 `embed` 功能将模板文件嵌入到编译后的二进制文件中。

### 分发方式

**只需要分发一个文件：**
- `git-log-analyzer` (或 `git-log-analyzer.exe` on Windows)

**不需要额外文件：**
- ❌ 不需要 `templates/` 目录
- ❌ 不需要任何 `.html`, `.css`, `.js` 文件
- ❌ 不需要配置文件（可选）

### 优势

1. **简化分发**：只需要一个二进制文件
2. **避免路径问题**：不会出现找不到模板文件的错误
3. **提高安全性**：模板内容无法被意外修改
4. **减少依赖**：用户无需维护额外的文件结构

### 技术实现

```go
//go:embed templates/report.html
var htmlTemplate string

//go:embed templates/styles.css
var cssTemplate string

//go:embed templates/charts.js
var jsTemplate string
```

模板文件在编译时被嵌入到二进制文件中，运行时直接从内存中读取。

### 编译要求

- Go 1.16+ (支持 embed 功能)
- 编译时 `templates/` 目录必须存在且包含必要文件

### 开发 vs 分发

- **开发时**：可以修改 `templates/` 目录中的文件，重新编译后生效
- **分发时**：只需要二进制文件，模板已嵌入其中
