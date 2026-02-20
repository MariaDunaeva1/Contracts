"""
Extractor Agent - Extract clauses from contract text
"""
from .base_agent import BaseAgent
from typing import Dict, Any


class ExtractorAgent(BaseAgent):
    """Extract key clauses from contract text"""
    
    def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Extract clauses from contract
        
        Args:
            input_data: Dict with 'contract_text' key
        
        Returns:
            Dict with 'clauses' list
        """
        contract_text = input_data.get('contract_text', '')
        
        if not contract_text:
            return {'clauses': [], 'error': 'No contract text provided'}
        
        prompt = f"""
You are a legal contract analyzer. Extract key clauses from this contract.

For each clause, identify:
1. Type (indemnification, payment, termination, liability, confidentiality, warranty, dispute_resolution, intellectual_property, or other)
2. Full text of the clause (keep it concise, max 200 words)
3. Risk level (high, medium, or low)
4. Brief reasoning for the risk level

Contract text:
{contract_text[:4000]}

Return ONLY valid JSON in this exact format (no markdown, no extra text):
{{
    "clauses": [
        {{
            "type": "indemnification",
            "text": "The full clause text here...",
            "risk_level": "high",
            "reasoning": "Why this is high risk..."
        }}
    ]
}}
"""
        
        response = self.complete(prompt, temperature=0.3)
        result = self.parse_response(response)
        
        # Validate and clean up
        if 'clauses' in result:
            # Ensure all clauses have required fields
            valid_clauses = []
            for clause in result['clauses']:
                if all(k in clause for k in ['type', 'text', 'risk_level']):
                    valid_clauses.append({
                        'type': clause['type'],
                        'text': clause['text'],
                        'risk_level': clause['risk_level'],
                        'reasoning': clause.get('reasoning', '')
                    })
            
            return {'clauses': valid_clauses}
        
        return result
