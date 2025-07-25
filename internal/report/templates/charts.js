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
                animation: {
                    duration: 800
                },
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

    // Files chart
    const filesCtx = document.getElementById('filesChart');
    if (filesCtx && data.files && data.files.length > 0) {
        new Chart(filesCtx, {
            type: 'bar',
            data: {
                labels: data.files.slice(0, 10).map(f => f.Name.split('/').pop()), // 只显示文件名
                datasets: [{
                    label: '修改次数',
                    data: data.files.slice(0, 10).map(f => f.Count),
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
                indexAxis: 'y', // 水平条形图
                scales: {
                    x: {
                        beginAtZero: true
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            title: function(context) {
                                const index = context[0].dataIndex;
                                return data.files[index].Name; // 显示完整路径
                            }
                        }
                    }
                }
            }
        });
    } else {
        console.warn('Files chart: No data or element not found');
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

// 提交森林图功能
function initCommitForest(branchData) {
    if (!branchData || !branchData.branches) {
        console.log('No branch data available for forest chart');
        return;
    }

    console.log('Initializing commit forest with data:', branchData);
    
    const svg = d3.select('#commitForest');
    if (svg.empty()) {
        console.error('Forest SVG element not found');
        return;
    }

    // 清空现有内容
    svg.selectAll('*').remove();

    // 设置尺寸和边距
    const margin = { top: 40, right: 40, bottom: 40, left: 100 };
    const width = 800 - margin.left - margin.right;
    const height = 500 - margin.top - margin.bottom;

    // 创建缩放行为
    const zoom = d3.zoom()
        .scaleExtent([0.5, 3])
        .on('zoom', function(event) {
            g.attr('transform', event.transform);
        });

    svg.call(zoom);

    // 创建主绘图组
    const g = svg.append('g')
        .attr('transform', `translate(${margin.left},${margin.top})`);

    // 处理分支数据
    const branches = branchData.branches || [];
    const commitGraph = branchData.commit_graph || [];
    
    if (commitGraph.length === 0) {
        g.append('text')
            .attr('x', width / 2)
            .attr('y', height / 2)
            .attr('text-anchor', 'middle')
            .text('暂无提交数据')
            .style('font-size', '16px')
            .style('fill', '#666');
        return;
    }

    // 按时间排序提交
    commitGraph.sort((a, b) => new Date(a.date) - new Date(b.date));

    // 为每个分支分配颜色和Y位置
    const branchIndexMap = {};
    branches.forEach((branch, index) => {
        branchIndexMap[branch.name] = index;
    });

    // 创建比例尺
    const xScale = d3.scaleTime()
        .domain(d3.extent(commitGraph, d => new Date(d.date)))
        .range([0, width]);

    const yScale = d3.scaleLinear()
        .domain([0, branches.length - 1])
        .range([height - 100, 50]);

    // 按分支绘制线条
    branches.forEach((branch, index) => {
        const branchCommits = commitGraph
            .filter(c => c.branch === branch.name)
            .map(c => ({
                date: new Date(c.date),
                branchIndex: index
            }));

        if (branchCommits.length === 0) return;

        // 创建线条生成器
        const line = d3.line()
            .x(d => xScale(d.date))
            .y(d => yScale(d.branchIndex))
            .curve(d3.curveMonotoneX);

        // 绘制分支线
        g.append('path')
            .datum(branchCommits)
            .attr('class', 'branch-line')
            .attr('d', line)
            .attr('stroke', getBranchColor(index))
            .attr('stroke-width', 3)
            .attr('fill', 'none')
            .attr('opacity', 0.7);

        // 绘制分支标签
        g.append('text')
            .attr('x', -10)
            .attr('y', yScale(index))
            .attr('dy', '0.35em')
            .attr('text-anchor', 'end')
            .text(branch.name)
            .style('font-size', '12px')
            .style('font-weight', 'bold')
            .style('fill', getBranchColor(index));
    });

    // 绘制提交节点
    const commitNodes = g.selectAll('.commit-node')
        .data(commitGraph)
        .enter().append('circle')
        .attr('class', 'commit-node')
        .attr('cx', d => xScale(new Date(d.date)))
        .attr('cy', d => yScale(branchIndexMap[d.branch] || 0))
        .attr('r', d => d.is_merge ? 6 : 4)
        .attr('fill', d => d.is_merge ? '#dc3545' : '#28a745')
        .attr('stroke', d => d.is_merge ? '#c82333' : '#1e7e34')
        .attr('stroke-width', 2)
        .on('mouseover', function(event, d) {
            showCommitTooltip(event, d);
        })
        .on('mouseout', hideCommitTooltip)
        .on('click', function(event, d) {
            showCommitDetails(d);
        });

    // 添加时间轴
    const xAxis = d3.axisBottom(xScale)
        .tickFormat(d3.timeFormat('%Y-%m-%d'));

    g.append('g')
        .attr('transform', `translate(0,${height - 30})`)
        .call(xAxis)
        .selectAll('text')
        .style('text-anchor', 'end')
        .attr('dx', '-.8em')
        .attr('dy', '.15em')
        .attr('transform', 'rotate(-45)');

    // 设置控件事件
    setupForestControls(svg, g, zoom, branches, commitGraph, branchIndexMap);
}

function getBranchColor(index) {
    const colors = [
        '#007bff', '#28a745', '#dc3545', '#ffc107', 
        '#17a2b8', '#6f42c1', '#fd7e14', '#20c997'
    ];
    return colors[index % colors.length];
}

function showCommitTooltip(event, commit) {
    const tooltip = d3.select('body').append('div')
        .attr('class', 'commit-tooltip')
        .style('position', 'absolute')
        .style('background', 'rgba(0, 0, 0, 0.8)')
        .style('color', 'white')
        .style('padding', '8px')
        .style('border-radius', '4px')
        .style('font-size', '12px')
        .style('pointer-events', 'none')
        .style('z-index', '1000');

    tooltip.html(`
        <div><strong>分支:</strong> ${commit.branch}</div>
        <div><strong>作者:</strong> ${commit.author}</div>
        <div><strong>时间:</strong> ${new Date(commit.date).toLocaleString()}</div>
        <div><strong>消息:</strong> ${commit.message || commit.short_hash}</div>
    `)
    .style('left', (event.pageX + 10) + 'px')
    .style('top', (event.pageY - 10) + 'px');
}

function hideCommitTooltip() {
    d3.selectAll('.commit-tooltip').remove();
}

function showCommitDetails(commit) {
    const infoPanel = document.getElementById('commitInfo');
    const detailsDiv = document.getElementById('commitDetails');
    
    if (infoPanel && detailsDiv) {
        detailsDiv.innerHTML = `
            <h4>提交详情</h4>
            <p><strong>分支:</strong> ${commit.branch}</p>
            <p><strong>哈希:</strong> ${commit.hash}</p>
            <p><strong>短哈希:</strong> ${commit.short_hash}</p>
            <p><strong>作者:</strong> ${commit.author}</p>
            <p><strong>时间:</strong> ${new Date(commit.date).toLocaleString()}</p>
            <p><strong>消息:</strong> ${commit.message || '无消息'}</p>
            <p><strong>合并提交:</strong> ${commit.is_merge ? '是' : '否'}</p>
            ${commit.parents && commit.parents.length > 0 ? 
                `<p><strong>父提交:</strong> ${commit.parents.join(', ')}</p>` : ''
            }
        `;
        infoPanel.style.display = 'block';
    }
}

function setupForestControls(svg, g, zoom, branches, commitGraph, branchIndexMap) {
    // 缩放控制
    document.getElementById('zoomIn')?.addEventListener('click', () => {
        svg.transition().call(zoom.scaleBy, 1.5);
    });

    document.getElementById('zoomOut')?.addEventListener('click', () => {
        svg.transition().call(zoom.scaleBy, 1 / 1.5);
    });

    document.getElementById('resetView')?.addEventListener('click', () => {
        svg.transition().call(zoom.transform, d3.zoomIdentity);
    });

    // 分支过滤
    const branchFilter = document.getElementById('branchFilter');
    if (branchFilter) {
        branchFilter.addEventListener('change', (e) => {
            const selectedBranch = e.target.value;
            filterBranchView(g, selectedBranch, branches, commitGraph, branchIndexMap);
        });
    }
}

function filterBranchView(g, selectedBranch, branches, commitGraph, branchIndexMap) {
    if (selectedBranch === 'all') {
        // 显示所有分支
        g.selectAll('.branch-line').style('opacity', 0.7);
        g.selectAll('.commit-node').style('opacity', 1);
        g.selectAll('text').style('opacity', 1);
    } else {
        // 只显示选中的分支
        const selectedIndex = branchIndexMap[selectedBranch];
        
        g.selectAll('.branch-line').style('opacity', (d, i) => i === selectedIndex ? 1 : 0.1);
        g.selectAll('.commit-node').style('opacity', d => d.branch === selectedBranch ? 1 : 0.1);
        g.selectAll('text').style('opacity', (d, i) => 
            branches[i] && branches[i].name === selectedBranch ? 1 : 0.3);
    }
}

// 确保D3.js库被加载
if (typeof d3 === 'undefined') {
    // 动态加载D3.js
    const script = document.createElement('script');
    script.src = 'https://d3js.org/d3.v7.min.js';
    script.onload = () => console.log('D3.js loaded successfully');
    script.onerror = () => console.error('Failed to load D3.js');
    document.head.appendChild(script);
}
