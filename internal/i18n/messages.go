package i18n

import (
	"os"
	"strings"
)

// Language represents supported languages
type Language string

const (
	LangZH Language = "zh" // Chinese
	LangEN Language = "en" // English
)

// Messages contains all translatable strings
type Messages struct {
	// Report titles
	ReportTitle             string
	TopContributors         string
	MostActiveHours         string
	MostModifiedFiles       string
	
	// Report fields
	TotalCommits            string
	ActivePeriod            string
	ActiveDays              string
	ActiveWeeks             string
	ActiveMonths            string
	Contributors            string
	CommitTimeline          string
	HourlyActivity          string
	DailyActivity           string
	CommitForest            string
	GeneratedOn             string
	
	// Units
	Commits                 string
	Lines                   string
	Modifications           string
	
	// Days of week
	DayNames                []string
	
	// AI Prompts
	AIPromptTemplate        string
	AISystemMessage         string
	AIAnalysisTitle         string
	AIKnowledgeBasePromptTemplate string
	AIKnowledgeBaseSystemMessage string
}

// translations contains all language translations
var translations = map[Language]*Messages{
	LangZH: {
		ReportTitle:             "Git 仓库分析报告",
		TopContributors:         "主要贡献者",
		MostActiveHours:         "最活跃时间",
		MostModifiedFiles:       "修改最多的文件",
		
		TotalCommits:            "总提交数",
		ActivePeriod:            "活跃周期",
		ActiveDays:              "活跃天数",
		ActiveWeeks:             "活跃周数",
		ActiveMonths:            "活跃月数",
		Contributors:            "贡献者",
		CommitTimeline:          "提交时间线",
		HourlyActivity:          "每小时活动",
		DailyActivity:           "每日活动",
		CommitForest:            "提交森林图",
		GeneratedOn:             "生成时间",
		
		Commits:                 "次提交",
		Lines:                   "行",
		Modifications:           "次修改",
		
		DayNames:                []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"},
		
		AIPromptTemplate: `请分析以下Git仓库统计数据并提供见解：

基础统计:
%s

详细数据:
- 总提交数: %d
- 贡献者数量: %d
- 活跃周期: %s 到 %s
- 活跃天数: %d

请提供:
1. 开发模式分析
2. 团队协作见解
3. 代码质量观察
4. 生产力趋势
5. 改进建议

重点关注可以帮助改进开发流程的可行见解。`,

		AISystemMessage: "你是一位专业的软件开发分析师。请分析Git仓库数据并提供可行的见解。请用中文回答。",
		AIAnalysisTitle: "智能分析",
		AIKnowledgeBasePromptTemplate: `请分析以下知识库项目的统计数据并提供见解：

基础统计:
%s

详细数据:
- 总编辑次数（文档更新）: %d
- 编辑者数量: %d
- 活跃周期: %s 到 %s
- 活跃天数: %d

请提供:
1. 知识库更新模式分析（内容热点和更新频率）
2. 编辑团队协作分布（谁在维护哪些部分）
3. 文档热点领域识别（最常修改的主题）
4. 知识库成熟度评估（内容完整度和稳定性）
5. 内容组织优化建议（改进知识库结构的建议）

特别关注：
- 文档质量和更新频率
- 内容的专业化和组织性
- 编辑团队的协作模式
- 需要补充或重写的内容领域`,
		AIKnowledgeBaseSystemMessage: "你是一位知识管理和文档架构专家。请分析知识库项目的编辑数据，重点关注文档的组织、更新频率、内容热点和知识结构的优化。提供实际可行的改进建议。请用中文回答。",
	},
	
	LangEN: {
		ReportTitle:             "Git Repository Analysis Report",
		TopContributors:         "Top Contributors",
		MostActiveHours:         "Most Active Hours",
		MostModifiedFiles:       "Most Modified Files",
		
		TotalCommits:            "Total Commits",
		ActivePeriod:            "Active Period",
		ActiveDays:              "Active Days",
		ActiveWeeks:             "Active Weeks",
		ActiveMonths:            "Active Months",
		Contributors:            "Contributors",
		CommitTimeline:          "Commit Timeline",
		HourlyActivity:          "Hourly Activity",
		DailyActivity:           "Daily Activity",
		CommitForest:            "Commit Forest",
		GeneratedOn:             "Generated on",
		
		Commits:                 "commits",
		Lines:                   "lines",
		Modifications:           "modifications",
		
		DayNames:                []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		
		AIPromptTemplate: `Please analyze the following Git repository statistics and provide insights:

BASIC STATISTICS:
%s

DETAILED DATA:
- Total commits: %d
- Number of contributors: %d
- Active period: %s to %s
- Active days: %d

Please provide:
1. Development pattern analysis
2. Team collaboration insights
3. Code quality observations
4. Productivity trends
5. Recommendations for improvement

Focus on actionable insights that can help improve the development process.`,

		AISystemMessage: "You are an expert software development analyst. Analyze git repository data and provide actionable insights.",
		AIAnalysisTitle: "Intelligent Analysis",
		AIKnowledgeBasePromptTemplate: `Please analyze the following knowledge base project statistics and provide insights:

BASIC STATISTICS:
%s

DETAILED DATA:
- Total document edits: %d
- Number of editors: %d
- Active period: %s to %s
- Active days: %d

Please provide:
1. Knowledge base update patterns analysis (content hotspots and update frequency)
2. Editor team collaboration distribution (who maintains which sections)
3. Document hotspot identification (most frequently modified topics)
4. Knowledge base maturity assessment (content completeness and stability)
5. Content organization optimization recommendations

Special focus on:
- Document quality and update frequency
- Content professionalization and organization
- Editor team collaboration patterns
- Content areas needing supplements or rewriting`,
		AIKnowledgeBaseSystemMessage: "You are a knowledge management and documentation architecture expert. Analyze knowledge base project editor data, focusing on document organization, update frequency, content hotspots, and knowledge structure optimization. Provide practical and actionable improvement recommendations.",
	},
}

// GetLanguage returns the current language setting
func GetLanguage() Language {
	lang := strings.ToLower(os.Getenv("REPORT_LANGUAGE"))
	switch lang {
	case "en", "english":
		return LangEN
	case "zh", "chinese", "zh-cn", "zh_cn":
		return LangZH
	default:
		return LangZH // Default to Chinese
	}
}

// GetMessages returns messages for the specified language
func GetMessages(lang Language) *Messages {
	if messages, exists := translations[lang]; exists {
		return messages
	}
	return translations[LangZH] // Fallback to Chinese
}

// T is a shorthand function for getting current language messages
func T() *Messages {
	return GetMessages(GetLanguage())
}
