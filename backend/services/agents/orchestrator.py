"""
Agent Orchestrator - Coordinate multi-agent workflow
"""
from .extractor_agent import ExtractorAgent
from .comparator_agent import ComparatorAgent
from .risk_agent import RiskAgent
from typing import Dict, Any
import uuid
import time


class AgentOrchestrator:
    """Orchestrate multi-agent workflow for contract analysis"""
    
    def __init__(self, llm_service, vector_service):
        """
        Initialize orchestrator with services
        
        Args:
            llm_service: LLMService instance
            vector_service: VectorService instance
        """
        self.llm = llm_service
        self.vector = vector_service
        
        # Agents will be initialized per request with model selection
        self.extractor = None
        self.comparator = None
        self.risk = None
    
    def analyze_contract(
        self, 
        contract_text: str, 
        contract_name: str = "New Contract",
        contract_id: str = None,
        use_finetuned: bool = True,
        knowledge_base_id: str = None
    ) -> Dict[str, Any]:
        """
        Full contract analysis workflow
        
        Args:
            contract_text: Full contract text
            contract_name: Human-readable name
            contract_id: Optional ID (generated if not provided)
            use_finetuned: Use fine-tuned model (True) or base model (False)
        
        Returns:
            Dict with complete analysis
        """
        start_time = time.time()
        
        # Generate ID if not provided
        if not contract_id:
            contract_id = f"contract_{uuid.uuid4().hex[:12]}"
        
        # Initialize agents with model selection
        self.extractor = ExtractorAgent(self.llm, use_finetuned=use_finetuned)
        self.comparator = ComparatorAgent(self.llm, use_finetuned=use_finetuned)
        self.risk = RiskAgent(self.llm, use_finetuned=use_finetuned)
        
        result = {
            'contract_id': contract_id,
            'contract_name': contract_name,
            'model_used': 'fine-tuned' if use_finetuned else 'base',
            'status': 'processing',
            'steps': {}
        }
        
        try:
            # Step 1: Extract clauses
            print(f"[Orchestrator] Step 1: Extracting clauses (model: {'fine-tuned' if use_finetuned else 'base'})...", flush=True)
            extraction_result = self.extractor.execute({
                'contract_text': contract_text
            })
            
            if 'error' in extraction_result:
                result['status'] = 'error'
                result['error'] = extraction_result['error']
                return result
            
            clauses = extraction_result.get('clauses', [])
            result['steps']['extraction'] = {
                'status': 'completed',
                'clauses_found': len(clauses)
            }
            
            if not clauses:
                result['status'] = 'completed'
                result['clauses'] = []
                result['comparisons'] = []
                result['risk_assessment'] = {
                    'overall_risk': 'unknown',
                    'message': 'No clauses extracted'
                }
                return result
            
            # Step 2: Index in vector DB
            print(f"[Orchestrator] Step 2: Indexing {len(clauses)} clauses...", flush=True)
            index_result = self.vector.index_contract(
                contract_id=contract_id,
                contract_name=contract_name,
                clauses=clauses
            )
            result['steps']['indexing'] = index_result
            
            # Step 3: Compare each clause with historical data
            print(f"[Orchestrator] Step 3: Comparing clauses...", flush=True)
            comparisons = []
            
            for i, clause in enumerate(clauses):
                print(f"  Analyzing clause {i+1}/{len(clauses)}: {clause['type']}", flush=True)
                
                # Find similar clauses (exclude current contract)
                # If knowledge_base_id is provided, filter by it using metadata.dataset_id or similar
                # Based on vector_service mapping, we likely want to filter by contract_id or a broad dataset tag if we have it.
                # Since indexing currently Uses contract_id, we'll implement filtering by KB ID in vector_service later if needed.
                # For now, we pass it down.
                filters = None
                if knowledge_base_id:
                    filters = {"contract_id": {"$ne": contract_id}} # Example filter logic
                
                similar = self.vector.search_similar_clauses(
                    query_text=clause['text'],
                    top_k=5,
                    filters=filters
                )
                
                # Filter out clauses from the same contract
                similar_results = [
                    r for r in similar.get('results', [])
                    if r['metadata']['contract_id'] != contract_id
                ][:3]  # Keep top 3
                
                # Compare if we have historical data
                if similar_results:
                    comparison = self.comparator.execute({
                        'clause': clause,
                        'similar_clauses': similar_results
                    })
                else:
                    comparison = {
                        'favorability_score': 0.0,
                        'comparison': 'No historical data available for comparison',
                        'key_differences': [],
                        'risks': [],
                        'recommendation': 'First contract of this type'
                    }
                
                comparisons.append({
                    'clause': clause,
                    'similar_clauses': similar_results,
                    'comparison': comparison
                })
            
            result['steps']['comparison'] = {
                'status': 'completed',
                'comparisons_made': len(comparisons)
            }
            
            # Step 4: Overall risk assessment
            print(f"[Orchestrator] Step 4: Assessing overall risk...", flush=True)
            risk_assessment = self.risk.execute({
                'comparisons': comparisons
            })
            result['steps']['risk_assessment'] = {
                'status': 'completed'
            }
            
            # Step 5: Generate summary
            summary = self.generate_summary(clauses, comparisons, risk_assessment)
            
            # Final result
            result['status'] = 'completed'
            result['clauses'] = clauses
            result['comparisons'] = comparisons
            result['risk_assessment'] = risk_assessment
            result['summary'] = summary
            result['processing_time'] = round(time.time() - start_time, 2)
            
            print(f"[Orchestrator] Analysis completed in {result['processing_time']}s", flush=True)
            
            return result
        
        except Exception as e:
            import traceback
            error_trace = traceback.format_exc()
            print(f"[Orchestrator] Error during analysis: {str(e)}", flush=True)
            print(f"[Orchestrator] Traceback: {error_trace}", flush=True)
            
            result['status'] = 'error'
            result['error'] = str(e)
            result['traceback'] = error_trace
            result['processing_time'] = round(time.time() - start_time, 2)
            return result
    
    def generate_summary(
        self, 
        clauses: list, 
        comparisons: list, 
        risk_assessment: dict
    ) -> Dict[str, Any]:
        """Generate executive summary"""
        
        high_risk_clauses = [
            c for c in clauses 
            if c.get('risk_level') == 'high'
        ]
        
        unfavorable_comparisons = [
            c for c in comparisons
            if c.get('comparison', {}).get('favorability_score', 0) < -0.3
        ]
        
        return {
            'total_clauses': len(clauses),
            'high_risk_count': len(high_risk_clauses),
            'unfavorable_count': len(unfavorable_comparisons),
            'overall_risk': risk_assessment.get('overall_risk', 'unknown'),
            'risk_score': risk_assessment.get('risk_score', 0.5),
            'key_findings': risk_assessment.get('top_risks', [])[:3],
            'executive_summary': risk_assessment.get('executive_summary', '')
        }
