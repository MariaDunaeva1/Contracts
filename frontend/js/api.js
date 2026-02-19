/**
 * API Client for Finetune Studio
 * Handles all backend communication
 */

// Use relative URL so it works with Nginx proxy
const API_BASE_URL = '/api/v1';

class API {
    constructor(baseUrl = API_BASE_URL) {
        this.baseUrl = baseUrl;
    }

    // Helper method for fetch requests
    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            ...options,
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                const error = await response.json().catch(() => ({ error: response.statusText }));
                throw new Error(error.error || `HTTP ${response.status}`);
            }

            // Handle empty responses
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }
            
            return response;
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    // Health check
    async health() {
        return this.request('/health');
    }

    // ========== DATASETS ==========

    async uploadDataset(file, name, description) {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('name', name);
        formData.append('description', description);

        const response = await fetch(`${this.baseUrl}/datasets`, {
            method: 'POST',
            body: formData,
        });

        if (!response.ok) {
            throw new Error(`Upload failed: ${response.statusText}`);
        }

        return response.json();
    }

    async listDatasets(page = 1, limit = 10) {
        return this.request(`/datasets?page=${page}&limit=${limit}`);
    }

    async getDataset(id) {
        return this.request(`/datasets/${id}`);
    }

    async deleteDataset(id) {
        return this.request(`/datasets/${id}`, { method: 'DELETE' });
    }

    // ========== JOBS ==========

    async createJob(datasetId, configuration) {
        return this.request('/jobs', {
            method: 'POST',
            body: JSON.stringify({
                dataset_id: datasetId,
                configuration: configuration,
            }),
        });
    }

    async listJobs(page = 1, limit = 10, status = null) {
        let url = `/jobs?page=${page}&limit=${limit}`;
        if (status) url += `&status=${status}`;
        return this.request(url);
    }

    async getJob(id) {
        return this.request(`/jobs/${id}`);
    }

    async cancelJob(id) {
        return this.request(`/jobs/${id}`, { method: 'DELETE' });
    }

    // Stream logs using SSE
    streamLogs(jobId, onMessage, onError) {
        const url = `${this.baseUrl}/jobs/${jobId}/logs`;
        const eventSource = new EventSource(url);

        eventSource.onmessage = (event) => {
            if (onMessage) onMessage(event.data);
        };

        eventSource.onerror = (error) => {
            console.error('SSE error:', error);
            if (onError) onError(error);
            eventSource.close();
        };

        return eventSource; // Return so caller can close it
    }

    async getLogs(jobId, limit = 100) {
        return this.request(`/jobs/${jobId}/logs?limit=${limit}`);
    }

    // ========== MODELS ==========

    async listModels(page = 1, limit = 10, filters = {}) {
        let url = `/models?page=${page}&limit=${limit}`;
        if (filters.base_model) url += `&base_model=${filters.base_model}`;
        if (filters.status) url += `&status=${filters.status}`;
        if (filters.date_from) url += `&date_from=${filters.date_from}`;
        if (filters.date_to) url += `&date_to=${filters.date_to}`;
        return this.request(url);
    }

    async getModel(id) {
        return this.request(`/models/${id}`);
    }

    async downloadModel(id) {
        window.open(`${this.baseUrl}/models/${id}/download`, '_blank');
    }

    async createModel(modelData) {
        return this.request('/models', {
            method: 'POST',
            body: JSON.stringify(modelData),
        });
    }

    async updateModel(id, updates) {
        return this.request(`/models/${id}`, {
            method: 'PUT',
            body: JSON.stringify(updates),
        });
    }

    async deleteModel(id) {
        return this.request(`/models/${id}`, { method: 'DELETE' });
    }

    // ========== EVALUATIONS ==========

    async createEvaluation(modelId, testSetPath, baseModelName) {
        return this.request(`/models/${modelId}/evaluate`, {
            method: 'POST',
            body: JSON.stringify({
                test_set_path: testSetPath,
                base_model_name: baseModelName,
            }),
        });
    }

    async listEvaluations(page = 1, limit = 10, filters = {}) {
        let url = `/evaluations?page=${page}&limit=${limit}`;
        if (filters.model_id) url += `&model_id=${filters.model_id}`;
        if (filters.status) url += `&status=${filters.status}`;
        return this.request(url);
    }

    async getEvaluation(id) {
        return this.request(`/evaluations/${id}`);
    }

    async updateEvaluation(id, updates) {
        return this.request(`/evaluations/${id}`, {
            method: 'PUT',
            body: JSON.stringify(updates),
        });
    }

    // ========== STATS ==========

    async getStats() {
        try {
            // Aggregate stats from multiple endpoints
            const [jobs, models, datasets] = await Promise.all([
                this.listJobs(1, 100).catch(() => ({ data: [], total: 0 })),
                this.listModels(1, 100).catch(() => ({ data: [], total: 0 })),
                this.listDatasets(1, 100).catch(() => ({ data: [], total: 0 })),
            ]);

            const completedJobs = (jobs.data || []).filter(j => j.status === 'completed');
            const totalTrainingTime = completedJobs.reduce((acc, job) => {
                const metrics = JSON.parse(job.metrics || '{}');
                return acc + (metrics.training_time || 0);
            }, 0);

            const successRate = jobs.total > 0 
                ? (completedJobs.length / jobs.total * 100).toFixed(1)
                : 0;

            return {
                totalJobs: jobs.total || 0,
                totalModels: models.total || 0,
                totalDatasets: datasets.total || 0,
                completedJobs: completedJobs.length,
                totalTrainingTime,
                successRate,
                recentJobs: (jobs.data || []).slice(0, 10),
            };
        } catch (error) {
            console.error('Error getting stats:', error);
            // Return default stats on error
            return {
                totalJobs: 0,
                totalModels: 0,
                totalDatasets: 0,
                completedJobs: 0,
                totalTrainingTime: 0,
                successRate: 0,
                recentJobs: [],
            };
        }
    }
}

// Export singleton instance
const api = new API();
