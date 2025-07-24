// AI Analysis Interactive Functions
function toggleAIAnalysis() {
    const content = document.getElementById('aiAnalysis-content');
    const icon = document.getElementById('aiAnalysis-icon');
    
    if (content.classList.contains('expanded')) {
        content.classList.remove('expanded');
        icon.classList.remove('expanded');
        icon.textContent = '▼';
    } else {
        content.classList.add('expanded');
        icon.classList.add('expanded');
        icon.textContent = '▲';
        
        // Render full AI analysis content
        renderAIAnalysisContent();
    }
}

function renderAIAnalysisContent() {
    const content = document.getElementById('aiAnalysis-content');
    if (content.querySelector('.ai-analysis-rendered')) {
        return; // Already rendered
    }
    
    const aiData = document.getElementById('aiAnalysisData');
    if (!aiData) {
        content.innerHTML = '<div class="ai-analysis-rendered"><p>暂无AI分析内容</p></div>';
        return;
    }
    
    // 直接获取文本内容，不需要JSON解析
    const analysisText = aiData.textContent || aiData.innerText;
    let htmlContent = '';
    
    if (analysisText && typeof marked !== 'undefined') {
        // 配置marked选项
        marked.setOptions({
            breaks: true,
            gfm: true,
            sanitize: false
        });
        
        htmlContent = marked.parse(analysisText);
    } else if (analysisText) {
        // 如果marked库未加载，使用简单的文本转换，保持换行
        htmlContent = '<pre style="white-space: pre-wrap; word-wrap: break-word;">' + 
                     analysisText.replace(/</g, '&lt;').replace(/>/g, '&gt;') + 
                     '</pre>';
    } else {
        htmlContent = '<p>暂无AI分析内容</p>';
    }
    
    content.innerHTML = '<div class="ai-analysis-rendered">' + htmlContent + '</div>';
}

function extractKeyInsights(analysisText, stats) {
    const insights = [];
    
    // 从统计数据中提取关键洞察
    if (stats && stats.authors) {
        const topContributor = stats.authors[0];
        if (topContributor) {
            insights.push({
                icon: '👨‍💻',
                title: '核心贡献者',
                value: `${topContributor.Name} (${topContributor.Percentage.toFixed(1)}%)`
            });
        }
    }
    
    if (stats && stats.timeline) {
        const maxCommitDay = stats.timeline.reduce((max, day) => 
            day.Count > max.Count ? day : max, stats.timeline[0]);
        insights.push({
            icon: '📈',
            title: '最活跃日期',
            value: maxCommitDay.Date
        });
    }
    
    // 从AI分析文本中提取更多洞察
    if (analysisText.includes('活跃') || analysisText.includes('频繁')) {
        insights.push({
            icon: '⚡',
            title: '开发节奏',
            value: '高频提交模式'
        });
    }
    
    if (analysisText.includes('协作') || analysisText.includes('团队')) {
        insights.push({
            icon: '🤝',
            title: '协作模式',
            value: '团队协作良好'
        });
    }
    
    return insights.slice(0, 4); // 最多显示4个洞察
}

function populateKeyInsights() {
    const aiData = document.getElementById('aiAnalysisData');
    const insightsContainer = document.getElementById('keyInsights');
    
    if (!aiData || !insightsContainer) return;
    
    // 直接获取文本内容，不需要JSON解析
    const analysisText = aiData.textContent || aiData.innerText;
    
    // Get chart data from global scope if available
    const chartData = window.chartData || {};
    
    const insights = extractKeyInsights(analysisText, chartData);
    
    insightsContainer.innerHTML = insights.map(insight => 
        `<div class="insight-card">
            <span class="insight-icon">${insight.icon}</span>
            <span class="insight-title">${insight.title}</span>
            <span class="insight-value">${insight.value}</span>
        </div>`
    ).join('');
}

function initCharts(data) {
    console.log('Initializing charts with data:', data);
    
    // Store chart data globally for AI analysis features
    window.chartData = data;
    
    // Validate data
    if (!data || !data.authors || !data.hourly || !data.daily || !data.timeline) {
        console.error('Invalid chart data provided:', data);
        return;
    }

    // Authors chart
    const authorsCtx = document.getElementById('authorsChart');
    if (authorsCtx && data.authors.length > 0) {
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
    } else {
        console.warn('Authors chart: No data or element not found');
    }

    // Timeline chart
    const timelineCtx = document.getElementById('timelineChart');
    if (timelineCtx && data.timeline.length > 0) {
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
    } else {
        console.warn('Timeline chart: No data or element not found');
    }

    // Hourly chart
    const hourlyCtx = document.getElementById('hourlyChart');
    if (hourlyCtx && data.hourly.length > 0) {
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
    } else {
        console.warn('Hourly chart: No data or element not found');
    }

    // Daily chart
    const dailyCtx = document.getElementById('dailyChart');
    if (dailyCtx && data.daily.length > 0) {
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
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
    } else {
        console.warn('Daily chart: No data or element not found');
    }
}

// Add error handling for Chart.js loading
document.addEventListener('DOMContentLoaded', function() {
    if (typeof Chart === 'undefined') {
        console.error('Chart.js library not loaded');
        document.querySelectorAll('.chart-container').forEach(container => {
            const canvas = container.querySelector('canvas');
            if (canvas) {
                canvas.style.display = 'none';
                const errorMsg = document.createElement('div');
                errorMsg.className = 'chart-error';
                errorMsg.textContent = 'Chart library not loaded';
                errorMsg.style.cssText = 'color: #999; text-align: center; padding: 20px;';
                container.appendChild(errorMsg);
            }
        });
    }
    
    // Initialize AI analysis features
    populateKeyInsights();
});
