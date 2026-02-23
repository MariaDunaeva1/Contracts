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

// Model complexity multipliers
const MODEL_SPECS = {
    'gpt2': { time: 0.1, vram: 0.5 },
    'qwen-2.5-1.5b': { time: 0.3, vram: 1.5 },
    'llama-3.2-1b': { time: 0.2, vram: 1.0 },
    'llama-3.2-3b': { time: 0.5, vram: 3.0 },
    'phi-3-mini': { time: 0.6, vram: 4.0 },
    'mistral-7b-v0.3': { time: 1.0, vram: 7.0 },
    'default': { time: 0.5, vram: 4.0 }
};

// Estimate training time based on configuration
function estimateTrainingTime(config) {
    const baseTime = 60 * 30; // 30 minutes base in seconds
    const modelSpec = MODEL_SPECS[config.model] || MODEL_SPECS['default'];

    const epochMultiplier = config.epochs || 3;
    const rankMultiplier = (config.lora_rank || 16) / 16;

    return Math.round(baseTime * modelSpec.time * epochMultiplier * rankMultiplier);
}

// Estimate VRAM usage
function estimateVRAM(config) {
    const modelSpec = MODEL_SPECS[config.model] || MODEL_SPECS['default'];
    const rankMultiplier = (config.lora_rank || 16) / 16;

    return Math.round(modelSpec.vram * rankMultiplier * 2); // Factor for training overhead
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
    const allowedExtensions = ['.json', '.jsonl', '.txt', '.csv', '.md', '.pdf', '.docx'];

    if (file.size > maxSize) {
        return { valid: false, error: 'File size exceeds 500MB limit' };
    }

    const fileName = file.name.toLowerCase();
    const hasValidExt = allowedExtensions.some(ext => fileName.endsWith(ext));

    if (!hasValidExt) {
        return { valid: false, error: 'Supported formats: JSON, JSONL, TXT, CSV, MD, PDF, DOCX' };
    }

    return { valid: true };
}

// Detect file format from extension
function getFileFormat(fileName) {
    const name = fileName.toLowerCase();
    if (name.endsWith('.json')) return 'json';
    if (name.endsWith('.jsonl')) return 'jsonl';
    if (name.endsWith('.csv')) return 'csv';
    if (name.endsWith('.md')) return 'markdown';
    if (name.endsWith('.txt')) return 'text';
    if (name.endsWith('.pdf')) return 'pdf';
    if (name.endsWith('.docx')) return 'docx';
    return 'unknown';
}

// Parse dataset preview efficiently without loading entire file into memory
async function parseDatasetPreview(file, maxLines = 10) {
    const format = getFileFormat(file.name);

    // PDF and DOCX cannot be previewed in the browser — parsing is server-side
    if (format === 'pdf' || format === 'docx') {
        return {
            format: format,
            total: 'Parsed on server',
            preview: [{ "info": `${format.toUpperCase()} document`, "message": "This file will be processed on the server. Preview not available in browser." }]
        };
    }

    return new Promise((resolve, reject) => {
        // Safety check: Don't even try to preview files over 50MB to prevent browser freeze
        if (file.size > 50 * 1024 * 1024) {
            resolve({
                format: format,
                total: 'Large file (>50MB)',
                preview: [{ "info": "Preview disabled", "message": "File is too large to preview in browser. Safe to upload." }]
            });
            return;
        }

        // Only read the first 1MB to avoid freezing the browser on large files
        const chunkSize = Math.min(1024 * 1024, file.size);
        const chunk = file.slice(0, chunkSize);

        const reader = new FileReader();

        reader.onload = (e) => {
            try {
                const text = e.target.result;
                const lines = text.split('\n').filter(line => line.trim());

                // For plain text formats (txt, csv, md), show raw lines
                if (format === 'text' || format === 'csv' || format === 'markdown') {
                    const previewLines = lines.slice(0, maxLines);
                    const totalLines = file.size <= chunkSize ? lines.length : 'Unknown (large file)';
                    resolve({
                        format: format,
                        total: totalLines,
                        preview: previewLines.map(line => ({ "text": line })),
                    });
                    return;
                }

                // Try to parse as JSON array (if it fits in the chunk or is complete)
                try {
                    if (file.size <= chunkSize) {
                        const data = JSON.parse(text);
                        if (Array.isArray(data) || data.messages) {
                            const arr = Array.isArray(data) ? data : data.messages;
                            resolve({
                                format: 'json',
                                total: arr.length,
                                preview: arr.slice(0, maxLines),
                            });
                            return;
                        }
                    }
                } catch (e) {
                    // Not a complete JSON array or too large, fallback to regex extraction for large JSON arrays
                }

                if (format === 'json') {
                    // Attempt to extract objects using basic RegExp if it's a large JSON array and JSON.parse failed
                    const preview = [];
                    const objRegex = /\{[^{}]*\}/g; // Basic regex to match simple non-nested objects or find the start
                    let match;
                    let matchesFound = 0;

                    // A more robust but simple way for preview is just to split by "},{" or "\n" if it's pretty-printed
                    // For a reliable preview, we can just send back raw text chunks
                    const linesPreview = lines.slice(0, maxLines * 5).join('\n');

                    resolve({
                        format: 'json',
                        total: 'Unknown (large file)',
                        preview: [{ "info": "Large JSON file", "preview_text": linesPreview.substring(0, 1000) + "..." }],
                    });
                    return;
                }

                // Parse as JSONL - only take complete lines (ignore last potentially cut line)
                const previewLines = lines.length > 1 ? lines.slice(0, lines.length - 1) : lines;
                const preview = [];

                for (let i = 0; i < Math.min(maxLines, previewLines.length); i++) {
                    try {
                        preview.push(JSON.parse(previewLines[i]));
                    } catch (err) {
                        // Skip invalid JSON lines that might be cut
                    }
                }

                resolve({
                    format: 'jsonl',
                    total: 'Unknown (large file)',
                    preview: preview,
                });
            } catch (error) {
                reject(new Error('Failed to parse file'));
            }
        };

        reader.onerror = () => reject(new Error('Failed to read file segment'));
        reader.readAsText(chunk);
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
