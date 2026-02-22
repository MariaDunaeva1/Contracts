/**
 * Contract Analysis with RAG
 */

// Base models available (Groq cloud API)
const BASE_MODELS = [
    { id: 'llama-3.1-8b-instant', name: 'Llama 3.1 8B Instant', type: 'base' },
    { id: 'llama-3.3-70b-versatile', name: 'Llama 3.3 70B Versatile', type: 'base' },
];

// Models from fine-tuning section (available for fine-tuning)
const FINETUNE_BASE_MODELS = [
    { id: 'llama-3.2-1b', name: '🦙 Llama 3.2 1B' },
    { id: 'qwen-2.5-1.5b', name: '🔴 Qwen 2.5 1.5B' },
    { id: 'llama-3.2-3b', name: '🦙 Llama 3.2 3B' },
    { id: 'phi-3-mini', name: '💎 Phi-3 Mini (3.8B)' },
    { id: 'mistral-7b-v0.3', name: '🌪️ Mistral 7B v0.3' },
    { id: 'gpt2', name: '📜 GPT-2 (124M)' },
];


function displayResults(result) {
    // Show results section
    document.getElementById('results').classList.add('active');

    // Display summary
    const riskAssessment = result.risk_assessment || {};
    const summary = result.summary || {};

    const riskBadge = document.getElementById('overallRisk');
    const overallRisk = riskAssessment.overall_risk || 'unknown';
    riskBadge.className = `risk-badge ${overallRisk}`;
    riskBadge.textContent = overallRisk.toUpperCase() + ' RISK';

    // Summary text
    const summaryText = document.getElementById('summaryText');
    summaryText.innerHTML = `
        <p><strong>Contract:</strong> ${result.contract_name || 'New Contract'}</p>
        <p><strong>Total Clauses:</strong> ${summary.total_clauses || 0}</p>
        <p><strong>High Risk Clauses:</strong> ${summary.high_risk_count || 0}</p>
        <p><strong>Unfavorable Comparisons:</strong> ${summary.unfavorable_count || 0}</p>
        <p><strong>Risk Score:</strong> ${((summary.risk_score || 0) * 100).toFixed(0)}%</p>
        <p><strong>Model:</strong> ${result.model_used || 'unknown'}</p>
        <p style="margin-top: 15px;">${riskAssessment.executive_summary || ''}</p>
    `;

    // Key findings
    const keyFindings = summary.key_findings || riskAssessment.top_risks || [];
    if (keyFindings.length > 0) {
        document.getElementById('keyFindings').style.display = 'block';
        const findingsList = document.getElementById('findingsList');
        findingsList.innerHTML = keyFindings.map(finding => `<li>${finding}</li>`).join('');
    }

    // Display clauses
    const comparisons = result.comparisons || [];
    const clausesList = document.getElementById('clausesList');

    if (comparisons.length === 0) {
        clausesList.innerHTML = '<p>No clauses extracted.</p>';
    } else {
        clausesList.innerHTML = comparisons.map((comp, index) => {
            const clause = comp.clause || {};
            const comparison = comp.comparison || {};
            const favorabilityScore = comparison.favorability_score || 0;

            return `
                <div class="clause-card ${clause.risk_level || 'medium'}">
                    <h3>${index + 1}. ${clause.type || 'Unknown'}</h3>
                    <p>${clause.text || ''}</p>
                    ${clause.reasoning ? `<p style="font-size: 13px; color: #888;"><em>${clause.reasoning}</em></p>` : ''}
                    <div class="favorability">
                        <strong>Favorability Score:</strong> ${favorabilityScore.toFixed(2)}
                        ${getFavorabilityBadge(favorabilityScore)}
                        <p style="margin-top: 10px; font-size: 14px;">${comparison.comparison || ''}</p>
                        ${comparison.risks && comparison.risks.length > 0 ? `
                            <p style="margin-top: 10px; font-size: 13px; color: #c62828;">
                                <strong>⚠️ Risks:</strong> ${comparison.risks.join('; ')}
                            </p>
                        ` : ''}
                        ${comparison.recommendation ? `
                            <p style="margin-top: 10px; font-size: 13px; color: #2e7d32;">
                                <strong>💡 Recommendation:</strong> ${comparison.recommendation}
                            </p>
                        ` : ''}
                    </div>
                </div>
            `;
        }).join('');
    }

    // Display similar clauses
    const similarList = document.getElementById('similarList');

    if (comparisons.length === 0) {
        similarList.innerHTML = '<p>No historical data available.</p>';
    } else {
        similarList.innerHTML = comparisons.map((comp, index) => {
            const clause = comp.clause || {};
            const similarClauses = comp.similar_clauses || [];

            if (similarClauses.length === 0) {
                return `
                    <div class="similar-group">
                        <h4>${index + 1}. Similar to: ${clause.type || 'Unknown'}</h4>
                        <p style="color: #888;">No historical clauses found for comparison.</p>
                    </div>
                `;
            }

            return `
                <div class="similar-group">
                    <h4>${index + 1}. Similar to: ${clause.type || 'Unknown'}</h4>
                    ${similarClauses.map(sim => {
                const metadata = sim.metadata || {};
                const similarity = (sim.similarity || 0) * 100;

                return `
                            <div class="similar-clause">
                                <span class="contract-name">${metadata.contract_name || 'Unknown Contract'}</span>
                                <span class="similarity">${similarity.toFixed(0)}% similar</span>
                                <div style="clear: both;"></div>
                                <span style="font-size: 12px; color: #888;">
                                    Type: ${metadata.clause_type || 'unknown'} | 
                                    Risk: ${metadata.risk_level || 'unknown'}
                                </span>
                                <p>${(sim.text || '').substring(0, 200)}${sim.text && sim.text.length > 200 ? '...' : ''}</p>
                            </div>
                        `;
            }).join('')}
                </div>
            `;
        }).join('');
    }

    // Scroll to results
    document.getElementById('results').scrollIntoView({ behavior: 'smooth' });
}

async function analyzeContract() {
    console.log('Starting analysis...');
    const contractText = document.getElementById('contractText').value.trim();
    const contractName = document.getElementById('contractName').value.trim() || 'New Contract';

    // Get selected model from dropdown
    const modelSelect = document.getElementById('modelSelect');
    const selectedModel = modelSelect.value;
    const kbId = document.getElementById('kbSelect').value;

    if (!contractText) {
        alert('Please enter contract text');
        return;
    }

    if (!selectedModel) {
        alert('Please select a model');
        return;
    }

    // Determine if this is a fine-tuned model or base model
    const selectedOption = modelSelect.selectedOptions[0];
    const modelType = selectedOption ? selectedOption.dataset.type : 'base';
    const useFinetuned = modelType === 'finetuned';

    // Show loading
    document.getElementById('loading').classList.add('active');
    document.getElementById('results').classList.remove('active');
    document.getElementById('analyzeBtn').disabled = true;

    try {
        console.log(`Calling API with model: ${selectedModel} (type: ${modelType})...`);
        const result = await api.analyzeContract({
            contract_text: contractText,
            contract_name: contractName,
            use_finetuned: useFinetuned,
            model_name: selectedModel,
            knowledge_base_id: kbId
        });

        console.log('Analysis result:', result);
        displayResults(result);

    } catch (error) {
        console.error('Analysis Error:', error);
        alert(`Analysis failed: ${error.message}\n\nMake sure the RAG service is running on port 8001.`);
    } finally {
        document.getElementById('loading').classList.remove('active');
        document.getElementById('analyzeBtn').disabled = false;
    }
}

function getFavorabilityBadge(score) {
    if (score > 0.3) {
        return '<span class="badge success">✓ Favorable</span>';
    } else if (score < -0.3) {
        return '<span class="badge danger">⚠️ Unfavorable</span>';
    } else {
        return '<span class="badge warning">~ Neutral</span>';
    }
}

// Load models into the dropdown
async function loadModels() {
    const select = document.getElementById('modelSelect');
    select.innerHTML = '';

    // Default option
    const defaultOpt = document.createElement('option');
    defaultOpt.value = '';
    defaultOpt.textContent = '-- Select Model --';
    select.appendChild(defaultOpt);

    // Group 1: Base models (Groq API)
    const baseGroup = document.createElement('optgroup');
    baseGroup.label = '🔵 Base Models (Groq API)';
    BASE_MODELS.forEach(model => {
        const opt = document.createElement('option');
        opt.value = model.id;
        opt.textContent = model.name;
        opt.dataset.type = 'base';
        baseGroup.appendChild(opt);
    });
    select.appendChild(baseGroup);

    // Group 2: Fine-tunable base models
    const finetuneBaseGroup = document.createElement('optgroup');
    finetuneBaseGroup.label = '🟡 Models Available for Fine-tuning';
    FINETUNE_BASE_MODELS.forEach(model => {
        const opt = document.createElement('option');
        opt.value = model.id;
        opt.textContent = model.name;
        opt.dataset.type = 'base';
        finetuneBaseGroup.appendChild(opt);
    });
    select.appendChild(finetuneBaseGroup);

    // Group 3: Fine-tuned models (from API)
    try {
        const modelsResponse = await api.listModels(1, 100, { status: 'ready' });
        if (modelsResponse.data && modelsResponse.data.length > 0) {
            const finetunedGroup = document.createElement('optgroup');
            finetunedGroup.label = '🟢 Fine-tuned Models';
            modelsResponse.data.forEach(model => {
                const opt = document.createElement('option');
                opt.value = model.name || `model-${model.ID}`;
                opt.textContent = `✨ ${model.name || model.base_model} (${model.base_model || 'custom'})`;
                opt.dataset.type = 'finetuned';
                opt.dataset.modelId = model.ID;
                finetunedGroup.appendChild(opt);
            });
            select.appendChild(finetunedGroup);
        }
    } catch (error) {
        console.warn('Could not load fine-tuned models:', error);
    }

    // Pre-select the versatile model
    select.value = 'llama-3.3-70b-versatile';
}

// Check RAG service health and load datasets/models on page load
window.addEventListener('DOMContentLoaded', async () => {
    loadKnowledgeBases();
    loadModels();

    try {
        const response = await fetch('/api/v1/rag/health');
        const health = await response.json();

        if (health.status !== 'healthy') {
            console.warn('RAG service is not fully healthy:', health);
            alert('Warning: RAG service may not be fully operational. Some features may not work.');
        } else {
            console.log('RAG service is healthy:', health);
        }
    } catch (error) {
        console.error('Failed to check RAG service health:', error);
        alert('Warning: Cannot connect to RAG service. Make sure it is running on port 8001.');
    }
});

async function loadKnowledgeBases() {
    try {
        const datasets = await api.listDatasets(1, 100);
        const select = document.getElementById('kbSelect');

        if (datasets.data && datasets.data.length > 0) {
            datasets.data.forEach(ds => {
                const option = document.createElement('option');
                option.value = ds.ID || ds.id; // Support both cases
                option.textContent = ds.name;
                select.appendChild(option);
            });
        }

        // Add LEDGAR as a permanent example if not already in the list
        const ledgarExists = datasets.data && datasets.data.some(ds => ds.name.toLowerCase().includes('ledgar'));
        if (!ledgarExists) {
            const ledgarOption = document.createElement('option');
            ledgarOption.value = "example-ledgar";
            ledgarOption.textContent = "🛡️ LEDGAR (Contract Clauses Example)";
            select.appendChild(ledgarOption);
        }
    } catch (error) {
        console.error('Failed to load knowledge bases:', error);
    }
}
