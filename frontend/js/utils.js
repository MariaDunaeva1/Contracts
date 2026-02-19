/**
 * Utility functions for Finetune Studio
 */

// Format time duration
function formatTime(seconds) {
    if (!seconds || seconds < 0) return '0s';
    
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    
    if (hours > 0) {
        return `${hours}h ${minutes}m`;
    } else if (minutes > 0) {
        return `${minutes}m ${secs}s`;
    } else {
        return `${secs}s`;
    }
}

// Format bytes to human readable
function formatBytes(bytes) {
    if (!bytes || bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Format date to relative time
function formatRelativeTime(date) {
    const now = new Date();
    const then = new Date(date);
    const diffMs = now - then;
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);
    
    if (diffSecs < 60) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    
    return then.toLocaleDateString();
}

// Format date to readable string
function formatDate(date) {
    return new Date(date).toLocaleString();
}

// Get status badge HTML
function getStatusBadge(status) {
    const badges = {
        pending: '<span class="badge badge-secondary">Pending</span>',
        starting: '<span class="badge badge-info">Starting</span>',
        running: '<span class="badge badge-primary">Running</span>',
        completed: '<span class="badge badge-success">Completed</span>',
        failed: '<span class="badge badge-danger">Failed</span>',
        cancelled: '<span class="badge badge-warning">Cancelled</span>',
        ready: '<span class="badge badge-success">Ready</span>',
        uploading: '<span class="badge badge-info">Uploading</span>',
        error: '<span class="badge badge-danger">Error</span>',
    };
    
    return badges[status] || `<span class="badge badge-secondary">${status}</span>`;
}

// Show toast notification
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.classList.add('show');
    }, 100);
    
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// Show loading spinner
function showLoading(element) {
    const spinner = document.createElement('div');
    spinner.className = 'spinner';
    spinner.innerHTML = '<div class="spinner-border" role="status"><span class="sr-only">Loading...</span></div>';
    element.appendChild(spinner);
    return spinner;
}

// Hide loading spinner
function hideLoading(spinner) {
    if (spinner && spinner.parentNode) {
        spinner.remove();
    }
}

// Debounce function
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Estimate training time based on configuration
function estimateTrainingTime(config) {
    const baseTime = 30; // 30 minutes base
    const epochMultiplier = config.epochs || 3;
    const rankMultiplier = (config.lora_rank || 16) / 16;
    
    const estimatedMinutes = baseTime * epochMultiplier * rankMultiplier;
    return Math.round(estimatedMinutes * 60); // Return in seconds
}

// Estimate VRAM usage
function estimateVRAM(config) {
    const baseVRAM = 4; // 4GB base
    const rankMultiplier = (config.lora_rank || 16) / 16;
    
    return Math.round(baseVRAM * rankMultiplier);
}

// Parse metrics from job
function parseMetrics(job) {
    try {
        return typeof job.metrics === 'string' 
            ? JSON.parse(job.metrics) 
            : job.metrics || {};
    } catch (e) {
        return {};
    }
}

// Parse configuration from job
function parseConfiguration(job) {
    try {
        return typeof job.configuration === 'string'
            ? JSON.parse(job.configuration)
            : job.configuration || {};
    } catch (e) {
        return {};
    }
}

// Calculate progress percentage
function calculateProgress(job) {
    const metrics = parseMetrics(job);
    
    if (job.status === 'completed') return 100;
    if (job.status === 'failed' || job.status === 'cancelled') return 0;
    if (job.status === 'pending') return 0;
    
    if (metrics.epoch && metrics.total_epochs) {
        return Math.round((metrics.epoch / metrics.total_epochs) * 100);
    }
    
    return 0;
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Copy to clipboard
async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        showToast('Copied to clipboard!', 'success');
    } catch (err) {
        console.error('Failed to copy:', err);
        showToast('Failed to copy', 'error');
    }
}

// Download JSON as file
function downloadJSON(data, filename) {
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

// Validate dataset file
function validateDatasetFile(file) {
    const maxSize = 500 * 1024 * 1024; // 500MB
    const allowedTypes = ['application/json', 'text/plain'];
    
    if (file.size > maxSize) {
        return { valid: false, error: 'File size exceeds 500MB limit' };
    }
    
    if (!allowedTypes.includes(file.type) && !file.name.endsWith('.json') && !file.name.endsWith('.jsonl')) {
        return { valid: false, error: 'Only JSON and JSONL files are allowed' };
    }
    
    return { valid: true };
}

// Parse dataset preview
async function parseDatasetPreview(file, maxLines = 10) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        
        reader.onload = (e) => {
            try {
                const text = e.target.result;
                const lines = text.split('\n').filter(line => line.trim());
                
                // Try to parse as JSON array
                try {
                    const data = JSON.parse(text);
                    if (Array.isArray(data)) {
                        resolve({
                            format: 'json',
                            total: data.length,
                            preview: data.slice(0, maxLines),
                        });
                        return;
                    }
                } catch (e) {
                    // Not a JSON array, try JSONL
                }
                
                // Parse as JSONL
                const preview = lines.slice(0, maxLines).map(line => JSON.parse(line));
                resolve({
                    format: 'jsonl',
                    total: lines.length,
                    preview: preview,
                });
            } catch (error) {
                reject(new Error('Invalid JSON format'));
            }
        };
        
        reader.onerror = () => reject(new Error('Failed to read file'));
        reader.readAsText(file);
    });
}

// Get color for metric value
function getMetricColor(value, inverse = false) {
    if (inverse) {
        if (value < 0.3) return 'success';
        if (value < 0.6) return 'warning';
        return 'danger';
    } else {
        if (value > 0.8) return 'success';
        if (value > 0.5) return 'warning';
        return 'danger';
    }
}
