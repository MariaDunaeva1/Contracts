# RAG + Agentes para Análisis de Contratos - Requirements

## Overview
Añadir sistema RAG (Retrieval-Augmented Generation) con ChromaDB para búsqueda semántica de contratos similares y agentes para análisis comparativo inteligente.

## User Stories

### 1. Como analista legal, quiero buscar cláusulas similares en contratos históricos
**Acceptance Criteria:**
- 1.1 Sistema indexa automáticamente contratos subidos
- 1.2 Búsqueda semántica encuentra cláusulas similares (no keyword matching)
- 1.3 Resultados incluyen metadata (tipo de cláusula, nivel de riesgo, contrato origen)
- 1.4 Top-K configurable (default: 5 resultados)
- 1.5 Búsqueda responde en < 500ms

### 2. Como analista legal, quiero comparar un nuevo contrato con históricos
**Acceptance Criteria:**
- 2.1 Upload de nuevo contrato trigger análisis automático
- 2.2 Fine-tuned model extrae cláusulas del nuevo contrato
- 2.3 Para cada cláusula, sistema busca similares en históricos
- 2.4 Comparación identifica términos más/menos favorables
- 2.5 Genera reporte con risk assessment

### 3. Como analista legal, quiero ver análisis comparativo visual
**Acceptance Criteria:**
- 3.1 UI muestra nuevo contrato con cláusulas extraídas
- 3.2 Cada cláusula tiene badge de riesgo (high/medium/low)
- 3.3 Panel lateral muestra cláusulas similares de contratos históricos
- 3.4 Comparación visual indica si términos son más/menos favorables
- 3.5 Summary ejecutivo del análisis

### 4. Como analista legal, quiero agentes que analicen automáticamente
**Acceptance Criteria:**
- 4.1 Agente extractor: identifica cláusulas clave
- 4.2 Agente comparador: compara con históricos
- 4.3 Agente de riesgo: evalúa nivel de riesgo
- 4.4 Agente de resumen: genera executive summary
- 4.5 Workflow orquestado automáticamente

### 5. Como desarrollador, quiero API endpoints para RAG
**Acceptance Criteria:**
- 5.1 POST /api/v1/contracts/analyze - Análisis completo con RAG
- 5.2 GET /api/v1/contracts/{id}/similar - Buscar similares
- 5.3 POST /api/v1/contracts/{id}/index - Indexar en vector DB
- 5.4 GET /api/v1/clauses/search - Búsqueda semántica de cláusulas
- 5.5 Todos los endpoints documentados

## Technical Requirements

### Vector Database
- ChromaDB con persistencia local
- Embeddings con sentence-transformers
- Colección "contracts" con metadata
- Índice optimizado para búsqueda rápida

### LLM Integration
- Usar modelo fine-tuned para extracción
- Ollama para inference local
- Prompts optimizados para comparación
- Streaming de respuestas largas

### Performance
- Indexación de contrato < 5 segundos
- Búsqueda semántica < 500ms
- Análisis completo < 30 segundos
- Soporte para contratos hasta 50 páginas

### Data Model
```python
Contract:
  - id: string
  - name: string
  - upload_date: datetime
  - clauses: [Clause]
  - indexed: boolean

Clause:
  - id: string
  - contract_id: string
  - text: string
  - type: string (indemnification, payment, termination, etc.)
  - risk_level: string (high, medium, low)
  - position: int

Comparison:
  - clause_id: string
  - similar_clauses: [SimilarClause]
  - analysis: string
  - favorability_score: float (-1 to 1)
  - risk_assessment: string
```

## Out of Scope
- OCR para contratos escaneados (solo texto)
- Multi-idioma (solo inglés)
- Integración con sistemas legales externos
- Firma electrónica
- Workflow de aprobación

## Success Metrics
- Búsqueda semántica encuentra cláusulas relevantes en >90% de casos
- Análisis completo en < 30 segundos
- UI responsive y fácil de usar
- Sistema indexa 100+ contratos sin degradación
