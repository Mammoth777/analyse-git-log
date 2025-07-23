package report

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"time"

	"git-log-analyzer/internal/analyzer"
)

// WebReportGenerator generates HTML reports
type WebReportGenerator struct {
	outputDir string
}

// NewWebReportGenerator creates a new web report generator
func NewWebReportGenerator(outputDir string) *WebReportGenerator {
	return &WebReportGenerator{
		outputDir: outputDir,
	}
}

// ReportData contains all data for web report
type ReportData struct {
	GeneratedAt      time.Time
	ProjectName      string
	Stats            *analyzer.Statistics
	TopAuthors       []AuthorData
	HourlyData       []HourData
	DailyData        []DayData
	FileData         []FileData
	CommitTimeline   []TimelineData
	AIAnalysis       string
}

// AuthorData represents author statistics for web display
type AuthorData struct {
	Name        string
	CommitCount int
	Additions   int
	Deletions   int
	Percentage  float64
}

// HourData represents hourly commit statistics
type HourData struct {
	Hour  int
	Count int
}

// DayData represents daily commit statistics
type DayData struct {
	Day   string
	Count int
}

// FileData represents file modification statistics
type FileData struct {
	Name  string
	Count int
}

// TimelineData represents commit timeline
type TimelineData struct {
	Date  string
	Count int
}

// GenerateReport generates a complete HTML report
func (w *WebReportGenerator) GenerateReport(stats *analyzer.Statistics, aiAnalysis string, projectName string) error {
	// Create output directory
	if err := os.MkdirAll(w.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Prepare report data
	reportData := w.prepareReportData(stats, aiAnalysis, projectName)

	// Generate HTML report
	if err := w.generateHTMLReport(reportData); err != nil {
		return err
	}

	// Generate CSS
	if err := w.generateCSS(); err != nil {
		return err
	}

	// Generate JavaScript
	if err := w.generateJavaScript(); err != nil {
		return err
	}

	return nil
}

// prepareReportData prepares data for web report
func (w *WebReportGenerator) prepareReportData(stats *analyzer.Statistics, aiAnalysis string, projectName string) *ReportData {
	data := &ReportData{
		GeneratedAt: time.Now(),
		ProjectName: projectName,
		Stats:       stats,
		AIAnalysis:  aiAnalysis,
	}

	// Prepare top authors
	type authorPair struct {
		key   string
		stats *analyzer.AuthorStat
	}
	
	var authors []authorPair
	for key, stat := range stats.AuthorStats {
		authors = append(authors, authorPair{key, stat})
	}
	
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].stats.CommitCount > authors[j].stats.CommitCount
	})

	for i, author := range authors {
		if i >= 10 { // Top 10 authors
			break
		}
		percentage := float64(author.stats.CommitCount) / float64(stats.TotalCommits) * 100
		data.TopAuthors = append(data.TopAuthors, AuthorData{
			Name:        author.stats.Name,
			CommitCount: author.stats.CommitCount,
			Additions:   author.stats.Additions,
			Deletions:   author.stats.Deletions,
			Percentage:  percentage,
		})
	}

	// Prepare hourly data
	for hour, count := range stats.TimeStats.HourlyPattern {
		data.HourlyData = append(data.HourlyData, HourData{
			Hour:  hour,
			Count: count,
		})
	}
	sort.Slice(data.HourlyData, func(i, j int) bool {
		return data.HourlyData[i].Hour < data.HourlyData[j].Hour
	})

	// Prepare daily data
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	for day, count := range stats.TimeStats.DailyPattern {
		data.DailyData = append(data.DailyData, DayData{
			Day:   dayNames[day],
			Count: count,
		})
	}

	// Prepare file data
	type filePair struct {
		file  string
		count int
	}
	
	var files []filePair
	for file, count := range stats.FileStats {
		files = append(files, filePair{file, count})
	}
	
	sort.Slice(files, func(i, j int) bool {
		return files[i].count > files[j].count
	})

	for i, f := range files {
		if i >= 15 { // Top 15 files
			break
		}
		data.FileData = append(data.FileData, FileData{
			Name:  f.file,
			Count: f.count,
		})
	}

	// Prepare timeline data
	type timelinePair struct {
		date  string
		count int
	}
	
	var timeline []timelinePair
	for date, count := range stats.CommitFrequency {
		timeline = append(timeline, timelinePair{date, count})
	}
	
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].date < timeline[j].date
	})

	for _, t := range timeline {
		data.CommitTimeline = append(data.CommitTimeline, TimelineData{
			Date:  t.date,
			Count: t.count,
		})
	}

	return data
}

// generateHTMLReport generates the main HTML report
func (w *WebReportGenerator) generateHTMLReport(data *ReportData) error {
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Git Log Analysis Report - {{.ProjectName}}</title>
    <link rel="stylesheet" href="styles.css">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="container">
        <header class="header">
            <h1>Git Log Analysis Report</h1>
            <div class="subtitle">
                <h2>{{.ProjectName}}</h2>
                <p>Generated on {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
            </div>
        </header>

        <div class="summary">
            <div class="stat-card">
                <h3>Total Commits</h3>
                <div class="stat-number">{{.Stats.TotalCommits}}</div>
            </div>
            <div class="stat-card">
                <h3>Contributors</h3>
                <div class="stat-number">{{len .Stats.AuthorStats}}</div>
            </div>
            <div class="stat-card">
                <h3>Active Days</h3>
                <div class="stat-number">{{.Stats.TimeStats.ActiveDays}}</div>
            </div>
            <div class="stat-card">
                <h3>Active Period</h3>
                <div class="stat-text">{{.Stats.TimeStats.FirstCommit.Format "2006-01-02"}} to {{.Stats.TimeStats.LastCommit.Format "2006-01-02"}}</div>
            </div>
        </div>

        <div class="charts-grid">
            <div class="chart-container">
                <h3>Top Contributors</h3>
                <canvas id="authorsChart"></canvas>
                <div class="authors-list">
                    {{range .TopAuthors}}
                    <div class="author-item">
                        <span class="author-name">{{.Name}}</span>
                        <span class="author-stats">{{.CommitCount}} commits ({{printf "%.1f" .Percentage}}%)</span>
                        <span class="author-changes">+{{.Additions}}/-{{.Deletions}}</span>
                    </div>
                    {{end}}
                </div>
            </div>

            <div class="chart-container">
                <h3>Commit Timeline</h3>
                <canvas id="timelineChart"></canvas>
            </div>

            <div class="chart-container">
                <h3>Hourly Activity</h3>
                <canvas id="hourlyChart"></canvas>
            </div>

            <div class="chart-container">
                <h3>Daily Activity</h3>
                <canvas id="dailyChart"></canvas>
            </div>
        </div>

        <div class="files-section">
            <h3>Most Modified Files</h3>
            <div class="files-grid">
                {{range .FileData}}
                <div class="file-item">
                    <span class="file-name">{{.Name}}</span>
                    <span class="file-count">{{.Count}} modifications</span>
                </div>
                {{end}}
            </div>
        </div>

        {{if .AIAnalysis}}
        <div class="ai-analysis">
            <h3>🤖 AI Analysis</h3>
            <div class="ai-content">
                <pre>{{.AIAnalysis}}</pre>
            </div>
        </div>
        {{end}}
    </div>

    <script src="charts.js"></script>
    <script>
        const reportData = {
            authors: {{.TopAuthors | printf "%+v"}},
            hourly: {{.HourlyData | printf "%+v"}},
            daily: {{.DailyData | printf "%+v"}},
            timeline: {{.CommitTimeline | printf "%+v"}}
        };
        initCharts(reportData);
    </script>
</body>
</html>`

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(w.outputDir, "index.html"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

// generateCSS generates the CSS file for the report
func (w *WebReportGenerator) generateCSS() error {
	css := `
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    line-height: 1.6;
    color: #333;
    background-color: #f5f7fa;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

.header {
    text-align: center;
    margin-bottom: 30px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 30px;
    border-radius: 10px;
    box-shadow: 0 4px 15px rgba(0,0,0,0.1);
}

.header h1 {
    font-size: 2.5em;
    margin-bottom: 10px;
}

.subtitle h2 {
    font-size: 1.5em;
    opacity: 0.9;
    margin-bottom: 5px;
}

.subtitle p {
    opacity: 0.8;
}

.summary {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 20px;
    margin-bottom: 30px;
}

.stat-card {
    background: white;
    padding: 25px;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    text-align: center;
    transition: transform 0.3s ease;
}

.stat-card:hover {
    transform: translateY(-5px);
}

.stat-card h3 {
    color: #666;
    font-size: 0.9em;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 10px;
}

.stat-number {
    font-size: 2.5em;
    font-weight: bold;
    color: #667eea;
}

.stat-text {
    font-size: 1.1em;
    color: #333;
    font-weight: 500;
}

.charts-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 30px;
    margin-bottom: 30px;
}

.chart-container {
    background: white;
    padding: 25px;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.chart-container h3 {
    margin-bottom: 20px;
    color: #333;
    font-size: 1.3em;
}

.chart-container canvas {
    max-height: 300px;
}

.authors-list {
    margin-top: 20px;
}

.author-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid #eee;
}

.author-item:last-child {
    border-bottom: none;
}

.author-name {
    font-weight: 500;
    color: #333;
}

.author-stats {
    color: #666;
}

.author-changes {
    font-family: monospace;
    color: #28a745;
    font-size: 0.9em;
}

.files-section {
    background: white;
    padding: 25px;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    margin-bottom: 30px;
}

.files-section h3 {
    margin-bottom: 20px;
    color: #333;
    font-size: 1.3em;
}

.files-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 15px;
}

.file-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 15px;
    background: #f8f9fa;
    border-radius: 5px;
    border-left: 4px solid #667eea;
}

.file-name {
    font-family: monospace;
    color: #333;
    font-weight: 500;
}

.file-count {
    color: #666;
    font-size: 0.9em;
}

.ai-analysis {
    background: white;
    padding: 25px;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    margin-bottom: 30px;
}

.ai-analysis h3 {
    margin-bottom: 20px;
    color: #333;
    font-size: 1.3em;
}

.ai-content {
    background: #f8f9fa;
    padding: 20px;
    border-radius: 5px;
    border-left: 4px solid #28a745;
}

.ai-content pre {
    white-space: pre-wrap;
    line-height: 1.6;
    color: #333;
}

@media (max-width: 768px) {
    .charts-grid {
        grid-template-columns: 1fr;
    }
    
    .summary {
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    }
    
    .files-grid {
        grid-template-columns: 1fr;
    }
    
    .author-item {
        flex-direction: column;
        align-items: flex-start;
        gap: 5px;
    }
}
`

	return os.WriteFile(filepath.Join(w.outputDir, "styles.css"), []byte(css), 0644)
}

// generateJavaScript generates the JavaScript file for charts
func (w *WebReportGenerator) generateJavaScript() error {
	js := `
function initCharts(data) {
    // Authors chart
    const authorsCtx = document.getElementById('authorsChart').getContext('2d');
    new Chart(authorsCtx, {
        type: 'doughnut',
        data: {
            labels: data.authors.map(a => a.Name),
            datasets: [{
                data: data.authors.map(a => a.CommitCount),
                backgroundColor: [
                    '#667eea', '#764ba2', '#f093fb', '#f5576c',
                    '#4facfe', '#00f2fe', '#43e97b', '#38f9d7',
                    '#ffecd2', '#fcb69f'
                ]
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'bottom'
                }
            }
        }
    });

    // Timeline chart
    const timelineCtx = document.getElementById('timelineChart').getContext('2d');
    new Chart(timelineCtx, {
        type: 'line',
        data: {
            labels: data.timeline.map(t => t.Date),
            datasets: [{
                label: 'Commits',
                data: data.timeline.map(t => t.Count),
                borderColor: '#667eea',
                backgroundColor: 'rgba(102, 126, 234, 0.1)',
                fill: true,
                tension: 0.4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });

    // Hourly chart
    const hourlyCtx = document.getElementById('hourlyChart').getContext('2d');
    new Chart(hourlyCtx, {
        type: 'bar',
        data: {
            labels: data.hourly.map(h => h.Hour + ':00'),
            datasets: [{
                label: 'Commits',
                data: data.hourly.map(h => h.Count),
                backgroundColor: 'rgba(102, 126, 234, 0.8)'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });

    // Daily chart
    const dailyCtx = document.getElementById('dailyChart').getContext('2d');
    new Chart(dailyCtx, {
        type: 'polarArea',
        data: {
            labels: data.daily.map(d => d.Day),
            datasets: [{
                data: data.daily.map(d => d.Count),
                backgroundColor: [
                    'rgba(255, 99, 132, 0.8)',
                    'rgba(54, 162, 235, 0.8)',
                    'rgba(255, 205, 86, 0.8)',
                    'rgba(75, 192, 192, 0.8)',
                    'rgba(153, 102, 255, 0.8)',
                    'rgba(255, 159, 64, 0.8)',
                    'rgba(199, 199, 199, 0.8)'
                ]
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false
        }
    });
}
`

	return os.WriteFile(filepath.Join(w.outputDir, "charts.js"), []byte(js), 0644)
}

// GetReportPath returns the path to the generated report
func (w *WebReportGenerator) GetReportPath() string {
	return filepath.Join(w.outputDir, "index.html")
}
