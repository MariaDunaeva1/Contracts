# RAG + Agentes - Design Document

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    RAG + Agents System                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐                     │
│  │   Frontend   │─────▶│  Comparison  │                     │
│  │     UI       │      │     API      │                     │
│  └──────────────┘      └──────┬───────┘                     │
│                               │                              │
│                               ▼                              │
│                    ┌─────────────────────┐                  │
│                    │  Agent Orchestrator │                  │
│                    └──────────┬──────────┘                  │
│                               │                              │
│         ┌─────────────────────┼─────────────────────┐       │
│         ▼                     ▼                     ▼        │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐ │
│  │  Extractor  │      │ Comparator  │      │    Risk     │ │
│  │   Agent     │      │   Agent     │      │   Agent     │ │
│  └──────┬──────┘      └──────┬──────┘      └──────┬──────┘ │
│         │                    │                     │         │
│         └────────────────────┼─────────────────────┘         │
│                              ▼                               │
│                    ┌─────────────────────┐                  │
│                    │   Vector Service    │                  │
│                    │    (ChromaDB)       │                  │
│                    └─────────────────────┘                  │
│                              │                               │
│                              ▼                               │
│                    ┌─────────────────────┐                  │
│                    │    LLM Service      │                  │
│                    │  (Fine-tuned Model) │                  │
│                    └─────────────────────┘                  │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. Vector Service (Python)

**File:** `backend/services/vector_service.py`

```python
from chromadb import Client
from chromadb.config import Settings
from sentence_transformers import SentenceTransformer
import uuid

class VectorService:
    def __init__(self, persist_dir="./chroma_db"):
        self.client = Client(Settings(
            chroma_db_impl="duckdb+parquet",
            persist_directory=persist_dir
        ))
        self.collection = self.client.get_or_create_collection(
            name="contracts",
            metadata={"hnsw:space": "cosine"}
        )
        # Use sentence-transformers for embeddings
        self.embedder = SentenceTransformer('all-MiniLM-L6-v2')
    
    def index_contract(self, contract_id: str, clauses: list) -> dict:
        """Index all clauses from a contract"""
        documents = []
        metadatas = []
        ids = []
        
        for i, clause in enumerate(clauses):
            clause_id = f"{contract_id}_clause_{i}"
            documents.append(clause['text'])
            metadatas.append({
                'contract_id': contract_id,
                'contract_name': clause.get('contract_name', ''),
                'clause_type': clause['type'],
                'risk_level': clause.get('risk_level', 'medium'),
                'position': i
            })
            ids.append(clause_id)
        
        self.collection.add(
            documents=documents,
            metadatas=metadatas,
            ids=ids
        )
        
        return {
            'contract_id': contract_id,
            'clauses_indexed': len(clauses),
            'status': 'success'
        }
    
    def search_similar_clauses(self, query_text: str, top_k: int = 5, 
                               filters: dict = None) -> dict:
        """Search for similar clauses"""
        where_clause = filters if filters else None
        
        results = self.collection.query(
            query_texts=[query_text],
            n_results=top_k,
            where=where_clause
        )
        
        return {
            'query': query_text,
            'results': [
                {
                    'text': doc,
                    'metadata': meta,
                    'distance': dist,
                    'similarity': 1 - dist  # Convert distance to similarity
                }
                for doc, meta, dist in zip(
                    results['documents'][0],
                    results['metadatas'][0],
                    results['distances'][0]
                )
            ]
        }
    
    def get_contract_clauses(self, contract_id: str) -> list:
        """Get all clauses for a contract"""
        results = self.collection.get(
            where={"contract_id": contract_id}
        )
        return results
    
    def delete_contract(self, contract_id: str):
        """Remove contract from index"""
        results = self.collection.get(
            where={"contract_id": contract_id}
        )
        if results['ids']:
            self.collection.delete(ids=results['ids'])
```

### 2. Agent System (Python)

**File:** `backend/services/agents/base_agent.py`

```python
from abc import ABC, abstractmethod
from typing import Dict, Any

class BaseAgent(ABC):
    def __init__(self, llm_service):
        self.llm = llm_service
    
    @abstractmethod
    def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """Execute agent task"""
        pass
```

**File:** `backend/services/agents/extractor_agent.py`

```python
from .base_agent import BaseAgent

class ExtractorAgent(BaseAgent):
    """Extract clauses from contract text"""
    
    def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        contract_text = input_data['contract_text']
        
        prompt = f"""
        Extract key clauses from this contract. For each clause, identify:
        1. Type (indemnification, payment, termination, liability, confidentiality, etc.)
        2. Full text of the clause
        3. Risk level (high, medium, low)
        
        Contract:
        {contract_text}
        
        Return JSON format:
        {{
            "clauses": [
                {{
                    "type": "indemnification",
                    "text": "...",
                    "risk_level": "high",
                    "reasoning": "..."
                }}
            ]
        }}
        """
        
        response = self.llm.complete(prompt)
        return self.parse_response(response)
    
    def parse_response(self, response: str) -> Dict[str, Any]:
        # Parse LLM response to structured format
        import json
        try:
            return json.loads(response)
        except:
            # Fallback parsing
            return {"clauses": []}
```

**File:** `backend/services/agents/comparator_agent.py`

```python
from .base_agent import BaseAgent

class ComparatorAgent(BaseAgent):
    """Compare clauses with similar historical ones"""
    
    def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        clause = input_data['clause']
        similar_clauses = input_data['similar_clauses']
        
        prompt = f"""
        Compare this new clause with similar clauses from past contracts.
        
        NEW CLAUSE:
        Type: {clause['type']}
        Text: {clause['text']}
        
        SIMILAR HISTORICAL CLAUSES:
        {self.format_similar(similar_clauses)}
        
        Analysis:
        1. Is this clause more or less favorable compared to historical ones?
        2. What are the key differences?
        3. What are potential risks or red flags?
        4. Favorability score (-1 to 1, where 1 is most favorable)
        
        Return JSON:
        {{
            "favorability_score": 0.5,
            "comparison": "...",
            "key_differences": ["...", "..."],
            "risks": ["...", "..."],
            "recommendation": "..."
        }}
        """
        
        response = self.llm.complete(prompt)
        return self.parse_response(response)
    
    def format_similar(self, similar_clauses: list) -> str:
        formatted = []
        for i, clause in enumerate(similar_clauses, 1):
            formatted.append(f"""
            {i}. Contract: {clause['metadata']['contract_name']}
               Similarity: {clause['similarity']:.2%}
               Text: {clause['text'][:200]}...
            """)
        return "\n".join(formatted)
```

**File:** `backend/services/agents/risk_agent.py`

```python
from .base_agent import BaseAgent

class RiskAgent(BaseAgent):
    """Assess overall contract risk"""
    
    def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        comparisons = input_data['comparisons']
        
        prompt = f"""
        Based on these clause comparisons, provide an overall risk assessment:
        
        {self.format_comparisons(comparisons)}
        
        Provide:
        1. Overall risk level (high, medium, low)
        2. Top 3 risk factors
        3. Recommended actions
        4. Executive summary
        
        Return JSON format.
        """
        
        response = self.llm.complete(prompt)
        return self.parse_response(response)
```

**File:** `backend/services/agents/orchestrator.py`

```python
from .extractor_agent import ExtractorAgent
from .comparator_agent import ComparatorAgent
from .risk_agent import RiskAgent

class AgentOrchestrator:
    """Orchestrate multi-agent workflow"""
    
    def __init__(self, llm_service, vector_service):
        self.llm = llm_service
        self.vector = vector_service
        self.extractor = ExtractorAgent(llm_service)
        self.comparator = ComparatorAgent(llm_service)
        self.risk = RiskAgent(llm_service)
    
    def analyze_contract(self, contract_text: str, contract_id: str) -> dict:
        """Full contract analysis workflow"""
        
        # Step 1: Extract clauses
        extraction_result = self.extractor.execute({
            'contract_text': contract_text
        })
        clauses = extraction_result['clauses']
        
        # Step 2: Index in vector DB
        self.vector.index_contract(contract_id, clauses)
        
        # Step 3: Compare each clause
        comparisons = []
        for clause in clauses:
            # Find similar clauses
            similar = self.vector.search_similar_clauses(
                clause['text'],
                top_k=3
            )
            
            # Compare
            comparison = self.comparator.execute({
                'clause': clause,
                'similar_clauses': similar['results']
            })
            
            comparisons.append({
                'clause': clause,
                'similar_clauses': similar['results'],
                'comparison': comparison
            })
        
        # Step 4: Risk assessment
        risk_assessment = self.risk.execute({
            'comparisons': comparisons
        })
        
        return {
            'contract_id': contract_id,
            'clauses': clauses,
            'comparisons': comparisons,
            'risk_assessment': risk_assessment,
            'summary': self.generate_summary(comparisons, risk_assessment)
        }
    
    def generate_summary(self, comparisons, risk_assessment) -> dict:
        high_risk_clauses = [
            c for c in comparisons 
            if c['clause']['risk_level'] == 'high'
        ]
        
        return {
            'total_clauses': len(comparisons),
            'high_risk_count': len(high_risk_clauses),
            'overall_risk': risk_assessment.get('overall_risk', 'medium'),
            'key_findings': risk_assessment.get('top_risks', [])
        }
```

### 3. Go API Handlers

**File:** `backend/internal/handlers/contract.go`

```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type ContractHandler struct {
    // Python service client
    pythonServiceURL string
}

func NewContractHandler(pythonURL string) *ContractHandler {
    return &ContractHandler{
        pythonServiceURL: pythonURL,
    }
}

// POST /api/v1/contracts/analyze
func (h *ContractHandler) AnalyzeContract(c *gin.Context) {
    var req struct {
        ContractText string `json:"contract_text" binding:"required"`
        ContractName string `json:"contract_name"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Call Python service
    result, err := h.callPythonService("/analyze", req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, result)
}

// GET /api/v1/contracts/:id/similar
func (h *ContractHandler) FindSimilar(c *gin.Context) {
    contractID := c.Param("id")
    topK := c.DefaultQuery("top_k", "5")
    
    result, err := h.callPythonService("/similar", map[string]interface{}{
        "contract_id": contractID,
        "top_k": topK,
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, result)
}

// POST /api/v1/clauses/search
func (h *ContractHandler) SearchClauses(c *gin.Context) {
    var req struct {
        Query string `json:"query" binding:"required"`
        TopK  int    `json:"top_k"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if req.TopK == 0 {
        req.TopK = 5
    }
    
    result, err := h.callPythonService("/search", req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, result)
}
```

### 4. Python FastAPI Service

**File:** `backend/services/rag_service.py`

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from vector_service import VectorService
from agents.orchestrator import AgentOrchestrator
from llm_service import LLMService
import uvicorn

app = FastAPI()

# Initialize services
vector_service = VectorService()
llm_service = LLMService()
orchestrator = AgentOrchestrator(llm_service, vector_service)

class AnalyzeRequest(BaseModel):
    contract_text: str
    contract_name: str = ""

class SearchRequest(BaseModel):
    query: str
    top_k: int = 5

@app.post("/analyze")
async def analyze_contract(req: AnalyzeRequest):
    try:
        result = orchestrator.analyze_contract(
            req.contract_text,
            contract_id=generate_id()
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/search")
async def search_clauses(req: SearchRequest):
    try:
        results = vector_service.search_similar_clauses(
            req.query,
            top_k=req.top_k
        )
        return results
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)
```

### 5. Frontend UI

**File:** `frontend/contract-analysis.html`

```html
<!DOCTYPE html>
<html>
<head>
    <title>Contract Analysis</title>
    <link rel="stylesheet" href="css/styles.css">
</head>
<body>
    <div class="container">
        <h1>Contract Analysis with RAG</h1>
        
        <!-- Upload Section -->
        <div class="upload-section">
            <textarea id="contractText" placeholder="Paste contract text here..."></textarea>
            <button onclick="analyzeContract()">Analyze Contract</button>
        </div>
        
        <!-- Results Section -->
        <div id="results" class="results-section" style="display:none;">
            <!-- Summary Card -->
            <div class="summary-card">
                <h2>Executive Summary</h2>
                <div class="risk-badge" id="overallRisk"></div>
                <div id="summaryText"></div>
            </div>
            
            <!-- Clauses Comparison -->
            <div class="comparison-grid">
                <div class="new-contract-panel">
                    <h2>Extracted Clauses</h2>
                    <div id="clausesList"></div>
                </div>
                
                <div class="similar-panel">
                    <h2>Similar Historical Clauses</h2>
                    <div id="similarList"></div>
                </div>
            </div>
        </div>
    </div>
    
    <script src="js/contract-analysis.js"></script>
</body>
</html>
```

**File:** `frontend/js/contract-analysis.js`

```javascript
async function analyzeContract() {
    const contractText = document.getElementById('contractText').value;
    
    if (!contractText) {
        alert('Please enter contract text');
        return;
    }
    
    showLoading();
    
    try {
        const response = await fetch('/api/v1/contracts/analyze', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                contract_text: contractText,
                contract_name: 'New Contract'
            })
        });
        
        const result = await response.json();
        displayResults(result);
    } catch (error) {
        console.error('Error:', error);
        alert('Analysis failed');
    }
}

function displayResults(result) {
    document.getElementById('results').style.display = 'block';
    
    // Display summary
    const riskBadge = document.getElementById('overallRisk');
    riskBadge.className = `risk-badge ${result.risk_assessment.overall_risk}`;
    riskBadge.textContent = result.risk_assessment.overall_risk.toUpperCase();
    
    // Display clauses
    const clausesList = document.getElementById('clausesList');
    clausesList.innerHTML = result.comparisons.map(comp => `
        <div class="clause-card ${comp.clause.risk_level}">
            <h3>${comp.clause.type}</h3>
            <p>${comp.clause.text}</p>
            <div class="favorability">
                Score: ${comp.comparison.favorability_score}
                ${getFavorabilityBadge(comp.comparison.favorability_score)}
            </div>
        </div>
    `).join('');
    
    // Display similar clauses
    const similarList = document.getElementById('similarList');
    similarList.innerHTML = result.comparisons.map(comp => `
        <div class="similar-group">
            <h4>Similar to: ${comp.clause.type}</h4>
            ${comp.similar_clauses.map(sim => `
                <div class="similar-clause">
                    <span class="contract-name">${sim.metadata.contract_name}</span>
                    <span class="similarity">${(sim.similarity * 100).toFixed(0)}% similar</span>
                    <p>${sim.text.substring(0, 150)}...</p>
                </div>
            `).join('')}
        </div>
    `).join('');
}

function getFavorabilityBadge(score) {
    if (score > 0.5) return '<span class="badge success">✓ Favorable</span>';
    if (score < -0.5) return '<span class="badge danger">⚠️ Unfavorable</span>';
    return '<span class="badge warning">~ Neutral</span>';
}
```

## Implementation Timeline

**Phase 1: Vector Service (4-5 hours)**
- Setup ChromaDB
- Implement VectorService
- Test indexing and search

**Phase 2: Agent System (6-8 hours)**
- Implement base agents
- Create orchestrator
- Test workflow

**Phase 3: API Integration (3-4 hours)**
- Python FastAPI service
- Go handlers
- Integration testing

**Phase 4: Frontend (4-5 hours)**
- Contract analysis UI
- Results visualization
- Interactive comparison

**Total: 17-22 hours**
