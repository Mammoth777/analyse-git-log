function initCharts(data) {
    console.log('Initializing charts with data:', data);
    
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
});
