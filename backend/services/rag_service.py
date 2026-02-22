"""
RAG Service - FastAPI service for contract analysis with RAG
"""
from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
import uvicorn
import uuid

from vector_service import VectorService
from llm_service import LLMService
from agents.orchestrator import AgentOrchestrator

# Initialize FastAPI app
app = FastAPI(
    title="LexAnalyzer - Contract Analysis API",
    description="Semantic search and AI-powered contract analysis",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize services
print("[LexAnalyzer] Initializing services...")
import os
vector_service = VectorService(persist_dir="./chroma_db")
provider = os.getenv("LLM_PROVIDER", "groq")  # Default to Groq
ollama_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
llm_service = LLMService(provider=provider, base_url=ollama_url)
orchestrator = AgentOrchestrator(llm_service, vector_service)
print(f"[LexAnalyzer] Services initialized successfully (Provider: {provider})")

# Request/Response models
class AnalyzeRequest(BaseModel):
    contract_text: str = Field(..., description="Full contract text")
    contract_name: str = Field(default="New Contract", description="Contract name")
    contract_id: Optional[str] = Field(default=None, description="Optional contract ID")
    knowledge_base_id: Optional[str] = Field(default=None, description="Optional Knowledge Base/Dataset ID for RAG filtering")
    use_finetuned: bool = Field(default=True, description="Use fine-tuned model (True) or base model (False)")
    model_name: Optional[str] = Field(default=None, description="Specific model name to use (overrides use_finetuned)")

class SearchRequest(BaseModel):
    query: str = Field(..., description="Search query text")
    top_k: int = Field(default=5, ge=1, le=20, description="Number of results")
    filters: Optional[Dict[str, str]] = Field(default=None, description="Metadata filters")

class IndexRequest(BaseModel):
    contract_id: str = Field(..., description="Contract ID")
    contract_name: str = Field(..., description="Contract name")
    clauses: List[Dict[str, Any]] = Field(..., description="List of clauses to index")

# Health check
@app.get("/health")
async def health_check():
    """Health check endpoint"""
    llm_available = llm_service.is_available()
    vector_stats = vector_service.get_stats()
    
    return {
        "status": "healthy" if llm_available else "degraded",
        "llm_service": "available" if llm_available else "unavailable",
        "vector_service": vector_stats.get('status', 'unknown'),
        "total_clauses_indexed": vector_stats.get('total_clauses', 0)
    }

# Main analysis endpoint
@app.post("/analyze")
async def analyze_contract(req: AnalyzeRequest):
    """
    Analyze contract with RAG
    
    This endpoint:
    1. Extracts clauses using fine-tuned model
    2. Indexes clauses in vector database
    3. Compares with historical contracts
    4. Provides risk assessment
    """
    try:
        model_info = req.model_name or ('fine-tuned' if req.use_finetuned else 'base')
        print(f"[API] Analyzing contract: {req.contract_name} (Model: {model_info})", flush=True)
        
        result = orchestrator.analyze_contract(
            contract_text=req.contract_text,
            contract_name=req.contract_name,
            contract_id=req.contract_id,
            use_finetuned=req.use_finetuned,
            knowledge_base_id=req.knowledge_base_id,
            model_name=req.model_name
        )
        
        if result.get('status') == 'error':
            raise HTTPException(status_code=500, detail=result.get('error', 'Analysis failed'))
        
        return result
    
    except Exception as e:
        print(f"[API] Error: {str(e)}", flush=True)
        raise HTTPException(status_code=500, detail=str(e))

# Semantic search endpoint
@app.post("/search")
async def search_clauses(req: SearchRequest):
    """
    Search for similar clauses using semantic search
    
    Returns clauses from historical contracts that are semantically similar
    to the query text.
    """
    try:
        print(f"[API] Searching for: {req.query[:50]}...")
        
        results = vector_service.search_similar_clauses(
            query_text=req.query,
            top_k=req.top_k,
            filters=req.filters
        )
        
        return results
    
    except Exception as e:
        print(f"[API] Error: {str(e)}", flush=True)
        raise HTTPException(status_code=500, detail=str(e))

# Index contract endpoint
@app.post("/index")
async def index_contract(req: IndexRequest):
    """
    Index a contract's clauses in the vector database
    
    Use this to manually index contracts without full analysis.
    """
    try:
        print(f"[API] Indexing contract: {req.contract_name}")
        
        result = vector_service.index_contract(
            contract_id=req.contract_id,
            contract_name=req.contract_name,
            clauses=req.clauses
        )
        
        if result.get('status') == 'error':
            raise HTTPException(status_code=500, detail=result.get('message', 'Indexing failed'))
        
        return result
    
    except Exception as e:
        print(f"[API] Error: {str(e)}", flush=True)
        raise HTTPException(status_code=500, detail=str(e))

# Get contract clauses
@app.get("/contracts/{contract_id}/clauses")
async def get_contract_clauses(contract_id: str):
    """Get all clauses for a specific contract"""
    try:
        result = vector_service.get_contract_clauses(contract_id)
        return result
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# Delete contract
@app.delete("/contracts/{contract_id}")
async def delete_contract(contract_id: str):
    """Remove contract from vector database"""
    try:
        result = vector_service.delete_contract(contract_id)
        return result
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# Get statistics
@app.get("/stats")
async def get_stats():
    """Get service statistics"""
    try:
        vector_stats = vector_service.get_stats()
        llm_models = llm_service.list_models()
        
        return {
            "vector_database": vector_stats,
            "llm_models": llm_models,
            "llm_available": llm_service.is_available()
        }
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/models")
async def list_analysis_models():
    """List available models for contract analysis"""
    try:
        llm_models = llm_service.list_models()
        return {
            "provider": llm_service.provider,
            "models": llm_models,
            "default_base": llm_service.model_base,
            "default_finetuned": llm_service.model_finetuned
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# Run server
if __name__ == "__main__":
    print("[LexAnalyzer] Starting server on http://0.0.0.0:8001")
    uvicorn.run(
        app, 
        host="0.0.0.0", 
        port=8001,
        log_level="info"
    )
