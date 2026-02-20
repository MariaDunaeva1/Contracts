/**
 * Contract Analysis with RAG
 */

async function analyzeContract() {
    const contractText = document.getElementById('contractText').value.trim();
    const contractName = document.getElementById('contractName').value.trim() || 'New Contract';
    const useFinetuned = document.getElementById('useFinetuned').checked;
    
    if (!contractText) {
        alert('Please enter contract text');
        return;
    }
    
    // Show loading
    document.getElementById('loading').classList.add('active');
    document.getElementById('results').classList.remove('active');
    document.getElementById('analyzeBtn').disabled = true;
    
    try {
        const response = await fetch('/api/v1/contracts/analyze', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                contract_text: contractText,
                contract_name: contractName,
                use_finetuned: useFinetuned
            })
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || error.details || 'Analysis failed');
        }
        
        const result = await response.json();
        console.log('Analysis result:', result);
        
        displayResults(result);
        
    } catch (error) {
        console.error('Error:', error);
        alert(`Analysis failed: ${error.message}\n\nMake sure the RAG service is running on port 8001.`);
    } finally {
        document.getElementById('loading').classList.remove('active');
        document.getElementById('analyzeBtn').disabled = false;
    }
}

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

function getFavorabilityBadge(score) {
    if (score > 0.3) {
        return '<span class="badge success">✓ Favorable</span>';
    } else if (score < -0.3) {
        return '<span class="badge danger">⚠️ Unfavorable</span>';
    } else {
        return '<span class="badge warning">~ Neutral</span>';
    }
}

// Check RAG service health on page load
window.addEventListener('DOMContentLoaded', async () => {
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
