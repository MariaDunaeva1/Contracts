/**
 * Chart utilities using Chart.js
 */

// Create loss curve chart
function createLossChart(canvasId, data) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return null;
    
    return new Chart(ctx, {
        type: 'line',
        data: {
            labels: data.epochs || [],
            datasets: [{
                label: 'Training Loss',
                data: data.loss || [],
                borderColor: 'rgb(75, 192, 192)',
                backgroundColor: 'rgba(75, 192, 192, 0.1)',
                tension: 0.4,
                fill: true,
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                },
                title: {
                    display: true,
                    text: 'Training Loss Over Time'
                }
            },
            scales: {
                y: {
                    beginAtZero: false,
                    title: {
                        display: true,
                        text: 'Loss'
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Epoch'
                    }
                }
            }
        }
    });
}

// Create accuracy chart
function createAccuracyChart(canvasId, data) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return null;
    
    return new Chart(ctx, {
        type: 'line',
        data: {
            labels: data.epochs || [],
            datasets: [{
                label: 'Accuracy',
                data: data.accuracy || [],
                borderColor: 'rgb(54, 162, 235)',
                backgroundColor: 'rgba(54, 162, 235, 0.1)',
                tension: 0.4,
                fill: true,
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                },
                title: {
                    display: true,
                    text: 'Accuracy Over Time'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 1,
                    title: {
                        display: true,
                        text: 'Accuracy'
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Epoch'
                    }
                }
            }
        }
    });
}

// Create comparison bar chart
function createComparisonChart(canvasId, baseMetrics, fineTunedMetrics) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return null;
    
    const labels = ['Accuracy', 'F1 Score', 'Precision', 'Recall'];
    const baseData = [
        baseMetrics.accuracy || 0,
        baseMetrics.f1_score || 0,
        baseMetrics.precision || 0,
        baseMetrics.recall || 0,
    ];
    const fineTunedData = [
        fineTunedMetrics.accuracy || 0,
        fineTunedMetrics.f1_score || 0,
        fineTunedMetrics.precision || 0,
        fineTunedMetrics.recall || 0,
    ];
    
    return new Chart(ctx, {
        type: 'bar',
        data: {
            labels: labels,
            datasets: [
                {
                    label: 'Base Model',
                    data: baseData,
                    backgroundColor: 'rgba(255, 159, 64, 0.7)',
                    borderColor: 'rgb(255, 159, 64)',
                    borderWidth: 1,
                },
                {
                    label: 'Fine-tuned Model',
                    data: fineTunedData,
                    backgroundColor: 'rgba(75, 192, 192, 0.7)',
                    borderColor: 'rgb(75, 192, 192)',
                    borderWidth: 1,
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                },
                title: {
                    display: true,
                    text: 'Model Performance Comparison'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 1,
                    title: {
                        display: true,
                        text: 'Score'
                    }
                }
            }
        }
    });
}

// Create dataset stats chart
function createDatasetStatsChart(canvasId, stats) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return null;
    
    return new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: Object.keys(stats),
            datasets: [{
                data: Object.values(stats),
                backgroundColor: [
                    'rgba(255, 99, 132, 0.7)',
                    'rgba(54, 162, 235, 0.7)',
                    'rgba(255, 206, 86, 0.7)',
                    'rgba(75, 192, 192, 0.7)',
                    'rgba(153, 102, 255, 0.7)',
                ],
                borderColor: [
                    'rgb(255, 99, 132)',
                    'rgb(54, 162, 235)',
                    'rgb(255, 206, 86)',
                    'rgb(75, 192, 192)',
                    'rgb(153, 102, 255)',
                ],
                borderWidth: 1,
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    position: 'right',
                },
                title: {
                    display: true,
                    text: 'Dataset Distribution'
                }
            }
        }
    });
}

// Create progress chart (radial)
function createProgressChart(canvasId, progress) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return null;
    
    return new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: ['Completed', 'Remaining'],
            datasets: [{
                data: [progress, 100 - progress],
                backgroundColor: [
                    'rgba(75, 192, 192, 0.7)',
                    'rgba(200, 200, 200, 0.3)',
                ],
                borderColor: [
                    'rgb(75, 192, 192)',
                    'rgb(200, 200, 200)',
                ],
                borderWidth: 1,
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            cutout: '70%',
            plugins: {
                legend: {
                    display: false,
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return context.label + ': ' + context.parsed + '%';
                        }
                    }
                }
            }
        }
    });
}

// Update chart data
function updateChartData(chart, newData) {
    if (!chart) return;
    
    chart.data.datasets[0].data = newData;
    chart.update();
}

// Add data point to chart
function addChartDataPoint(chart, label, data) {
    if (!chart) return;
    
    chart.data.labels.push(label);
    chart.data.datasets.forEach((dataset, index) => {
        dataset.data.push(data[index]);
    });
    chart.update();
}

// Destroy chart
function destroyChart(chart) {
    if (chart) {
        chart.destroy();
    }
}
