"""
Comparator Agent - Compare clauses with historical ones
"""
from .base_agent import BaseAgent
from typing import Dict, Any, List


class ComparatorAgent(BaseAgent):
    """Compare clauses with similar historical ones"""
    
    def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Compare clause with similar historical clauses
        
        Args:
            input_data: Dict with 'clause' and 'similar_clauses'
        
        Returns:
            Dict with comparison analysis
        """
        clause = input_data.get('clause', {})
        similar_clauses = input_data.get('similar_clauses', [])
        
        if not clause or not similar_clauses:
            return {
                'favorability_score': 0.0,
                'comparison': 'No similar clauses found for comparison',
                'key_differences': [],
                'risks': [],
                'recommendation': 'Unable to compare'
            }
        
        prompt = f"""
You are a legal contract analyst. Compare this new clause with similar clauses from past contracts.

NEW CLAUSE:
Type: {clause.get('type', 'unknown')}
Text: {clause.get('text', '')}
Current Risk Level: {clause.get('risk_level', 'unknown')}

SIMILAR HISTORICAL CLAUSES:
{self.format_similar(similar_clauses)}

Provide analysis in JSON format:
1. favorability_score: Float from -1 to 1 where:
   - 1.0 = Very favorable (better than historical)
   - 0.0 = Neutral (similar to historical)
   - -1.0 = Very unfavorable (worse than historical)

2. comparison: Brief text comparing the clauses (2-3 sentences)

3. key_differences: List of 2-3 main differences

4. risks: List of 1-3 potential risks or red flags

5. recommendation: Brief recommendation (1 sentence)

Return ONLY valid JSON (no markdown):
{{
    "favorability_score": 0.5,
    "comparison": "...",
    "key_differences": ["...", "..."],
    "risks": ["...", "..."],
    "recommendation": "..."
}}
"""
        
        response = self.complete(prompt, temperature=0.4)
        result = self.parse_response(response)
        
        # Ensure required fields exist
        if 'favorability_score' not in result:
            result['favorability_score'] = 0.0
        if 'comparison' not in result:
            result['comparison'] = 'Analysis unavailable'
        if 'key_differences' not in result:
            result['key_differences'] = []
        if 'risks' not in result:
            result['risks'] = []
        if 'recommendation' not in result:
            result['recommendation'] = 'Review carefully'
        
        return result
    
    def format_similar(self, similar_clauses: List[Dict]) -> str:
        """Format similar clauses for prompt"""
        if not similar_clauses:
            return "No similar clauses found."
        
        formatted = []
        for i, clause in enumerate(similar_clauses[:3], 1):  # Max 3 for context
            metadata = clause.get('metadata', {})
            text = clause.get('text', '')
            similarity = clause.get('similarity', 0)
            
            formatted.append(f"""
{i}. Contract: {metadata.get('contract_name', 'Unknown')}
   Similarity: {similarity:.1%}
   Type: {metadata.get('clause_type', 'unknown')}
   Risk Level: {metadata.get('risk_level', 'unknown')}
   Text: {text[:300]}{'...' if len(text) > 300 else ''}
""")
        
        return "\n".join(formatted)
