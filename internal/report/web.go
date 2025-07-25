package report

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"git-log-analyzer/internal/analyzer"
	"git-log-analyzer/internal/developer"
	"git-log-analyzer/internal/health"
	"git-log-analyzer/internal/i18n"
)

//go:embed templates/report.html
var htmlTemplate string

//go:embed templates/developer-profile.html
var developerProfileTemplate string

//go:embed templates/styles.css
var cssTemplate string

//go:embed templates/charts.js
var jsTemplate string

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
	GeneratedAt         time.Time
	ProjectName         string
	Stats               *analyzer.Statistics
	TopAuthors          []AuthorData
	HourlyData          []HourData
	DailyData           []DayData
	FileData            []FileData
	CommitTimeline      []TimelineData
	AIAnalysis          string
	AIStatus            AIStatus
	CodeHealthMetrics   *health.CodeHealthMetrics
	DeveloperProfiles   []*developer.DeveloperProfile
	Messages            *i18n.Messages
	Language            i18n.Language
}

// AIStatus represents the status of AI analysis
type AIStatus struct {
	Enabled       bool
	Available     bool
	ErrorType     string // "disabled", "config_error", "analysis_error"
	ErrorMessage  string
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
func (w *WebReportGenerator) GenerateReport(stats *analyzer.Statistics, aiAnalysis string, aiStatus AIStatus, projectName string, developerProfiles []*developer.DeveloperProfile) error {
	// Create output directory
	if err := os.MkdirAll(w.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Prepare report data
	reportData := w.prepareReportData(stats, aiAnalysis, aiStatus, projectName, developerProfiles)

	// Generate HTML report
	if err := w.generateHTMLReport(reportData); err != nil {
		return err
	}

	// Generate developer profile pages
	if err := w.generateDeveloperProfilePages(reportData); err != nil {
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
func (w *WebReportGenerator) prepareReportData(stats *analyzer.Statistics, aiAnalysis string, aiStatus AIStatus, projectName string, developerProfiles []*developer.DeveloperProfile) *ReportData {
	lang := i18n.GetLanguage()
	msg := i18n.GetMessages(lang)
	
	data := &ReportData{
		GeneratedAt:       time.Now(),
		ProjectName:       projectName,
		Stats:             stats,
		AIAnalysis:        aiAnalysis,
		AIStatus:          aiStatus,
		CodeHealthMetrics: stats.CodeHealthMetrics,
		DeveloperProfiles: developerProfiles,
		Messages:          msg,
		Language:          lang,
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
	for day, count := range stats.TimeStats.DailyPattern {
		data.DailyData = append(data.DailyData, DayData{
			Day:   msg.DayNames[day],
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
	// Create template functions for JSON serialization
	funcMap := template.FuncMap{
		"toJSON": func(v interface{}) template.JS {
			bytes, err := json.Marshal(v)
			if err != nil {
				return template.JS("{}")
			}
			return template.JS(string(bytes))
		},
		"slice": func(items interface{}, start, end int) interface{} {
			switch v := items.(type) {
			case []health.TechnicalDebtHotspot:
				if end > len(v) {
					end = len(v)
				}
				return v[start:end]
			case []health.RefactoringSignal:
				if end > len(v) {
					end = len(v)
				}
				return v[start:end]
			case []health.CodeConcentrationIssue:
				if end > len(v) {
					end = len(v)
				}
				return v[start:end]
			case []health.StabilityIndicator:
				if end > len(v) {
					end = len(v)
				}
				return v[start:end]
			default:
				return items
			}
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"printf": func(format string, v ...interface{}) string {
			return fmt.Sprintf(format, v...)
		},
		"getDeveloperProfileLink": func(authorName string, profiles []*developer.DeveloperProfile) string {
			for _, profile := range profiles {
				if profile.Name == authorName {
					return fmt.Sprintf("developer-%s.html", sanitizeFilename(profile.Name))
				}
			}
			return ""
		},
		"hasDeveloperProfile": func(authorName string, profiles []*developer.DeveloperProfile) bool {
			for _, profile := range profiles {
				if profile.Name == authorName {
					return true
				}
			}
			return false
		},
	}

	// Read HTML template from embedded content
	t, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
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

// generateDeveloperProfilePages generates individual developer profile pages
func (w *WebReportGenerator) generateDeveloperProfilePages(data *ReportData) error {
	if len(data.DeveloperProfiles) == 0 {
		return nil
	}

	// Create template functions
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"printf": func(format string, v ...interface{}) string {
			return fmt.Sprintf(format, v...)
		},
	}

	// Parse developer profile template
	t, err := template.New("developer-profile").Funcs(funcMap).Parse(developerProfileTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse developer profile template: %v", err)
	}

	// Generate a page for each developer profile
	for _, profile := range data.DeveloperProfiles {
		// Create profile data structure
		profileData := struct {
			DeveloperProfile *developer.DeveloperProfile
			Language         i18n.Language
		}{
			DeveloperProfile: profile,
			Language:         data.Language,
		}

		// Create filename based on developer name (sanitized)
		filename := fmt.Sprintf("developer-%s.html", sanitizeFilename(profile.Name))
		filePath := filepath.Join(w.outputDir, filename)

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create developer profile file %s: %v", filename, err)
		}

		err = t.Execute(file, profileData)
		file.Close()

		if err != nil {
			return fmt.Errorf("failed to generate developer profile for %s: %v", profile.Name, err)
		}
	}

	return nil
}

// sanitizeFilename sanitizes a string to be safe for use as a filename
func sanitizeFilename(name string) string {
	// Replace common problematic characters
	result := name
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "@", "-at-")
	result = strings.ReplaceAll(result, "/", "-")
	result = strings.ReplaceAll(result, "\\", "-")
	result = strings.ReplaceAll(result, ":", "-")
	result = strings.ReplaceAll(result, "*", "-")
	result = strings.ReplaceAll(result, "?", "-")
	result = strings.ReplaceAll(result, "\"", "-")
	result = strings.ReplaceAll(result, "<", "-")
	result = strings.ReplaceAll(result, ">", "-")
	result = strings.ReplaceAll(result, "|", "-")
	return result
}

// generateCSS generates the CSS file for the report
func (w *WebReportGenerator) generateCSS() error {
	// Use embedded CSS template
	return os.WriteFile(filepath.Join(w.outputDir, "styles.css"), []byte(cssTemplate), 0644)
}

// generateJavaScript generates the JavaScript file for charts
func (w *WebReportGenerator) generateJavaScript() error {
	// Use embedded JavaScript template
	return os.WriteFile(filepath.Join(w.outputDir, "charts.js"), []byte(jsTemplate), 0644)
}

// GetReportPath returns the path to the generated report
func (w *WebReportGenerator) GetReportPath() string {
	return filepath.Join(w.outputDir, "index.html")
}
