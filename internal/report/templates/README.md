# Report Templates

这个目录包含了Web报告生成所需的模板文件。

## 文件结构

- `report.html` - HTML模板文件，包含报告的结构和布局
- `styles.css` - CSS样式文件，定义报告的视觉样式
- `charts.js` - JavaScript文件，处理图表的初始化和渲染

## 模板特性

### HTML模板 (`report.html`)
- 使用Go模板语法 `{{.Variable}}`
- 支持条件渲染 `{{if .Condition}}`
- 支持循环 `{{range .Array}}`
- 集成Chart.js用于图表显示

### CSS样式 (`styles.css`)
- 响应式设计，支持移动设备
- 现代化的卡片式布局
- AI分析区域的特殊高亮样式
- 支持暗色和亮色主题元素

### JavaScript (`charts.js`)
- 使用Chart.js渲染各种图表
- 包含错误处理机制
- 支持多种图表类型：饼图、折线图、柱状图、极坐标图

## 自定义

要自定义报告的外观，您可以：

1. 修改 `styles.css` 来改变颜色、字体、布局
2. 编辑 `report.html` 来调整结构和内容
3. 更新 `charts.js` 来添加新的图表类型或修改现有图表

## 注意事项

- 模板文件使用UTF-8编码
- HTML模板必须包含有效的Go模板语法
- CSS文件包含媒体查询用于响应式设计
- JavaScript文件依赖Chart.js CDN
