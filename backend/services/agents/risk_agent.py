"""
Risk Agent - Assess overall contract risk
"""
from .base_agent import BaseAgent
from typing import Dict, Any, List


class RiskAgent(BaseAgent):
    """Assess overall contract risk based on clause comparisons"""
    
    def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Assess overall contract risk
        
        Args:
            input_data: Dict with 'comparisons' list
        
        Returns:
            Dict with risk assessment
        """
        comparisons = input_data.get('comparisons', [])
        
        if not comparisons:
            return {
                'overall_risk': 'unknown',
                'risk_score': 0.5,
                'top_risks': [],
                'recommendations': [],
                'executive_summary': 'No data available for risk assessment'
            }
        
        # Calculate basic metrics
        high_risk_count = sum(
            1 for c in comparisons 
            if c.get('clause', {}).get('risk_level') == 'high'
        )
        
        avg_favorability = sum(
            c.get('comparison', {}).get('favorability_score', 0)
            for c in comparisons
        ) / len(comparisons) if comparisons else 0
        
        prompt = f"""
You are a senior legal risk analyst. Assess the overall risk of this contract based on clause-by-clause analysis.

ANALYSIS SUMMARY:
- Total clauses analyzed: {len(comparisons)}
- High-risk clauses: {high_risk_count}
- Average favorability score: {avg_favorability:.2f} (scale: -1 to 1)

DETAILED CLAUSE ANALYSIS:
{self.format_comparisons(comparisons)}

Provide comprehensive risk assessment in JSON format:

1. overall_risk: "high", "medium", or "low"
2. risk_score: Float from 0 to 1 (0=low risk, 1=high risk)
3. top_risks: List of 3-5 most critical risk factors
4. recommendations: List of 3-5 recommended actions
5. executive_summary: 2-3 sentence summary for executives

Return ONLY valid JSON (no markdown):
{{
    "overall_risk": "medium",
    "risk_score": 0.6,
    "top_risks": ["...", "...", "..."],
    "recommendations": ["...", "...", "..."],
    "executive_summary": "..."
}}
"""
        
        response = self.complete(prompt, temperature=0.3, max_tokens=1500)
        result = self.parse_response(response)
        
        # Ensure required fields
        if 'overall_risk' not in result:
            result['overall_risk'] = self.calculate_risk_level(high_risk_count, len(comparisons))
        if 'risk_score' not in result:
            result['risk_score'] = self.calculate_risk_score(high_risk_count, len(comparisons), avg_favorability)
        if 'top_risks' not in result:
            result['top_risks'] = []
        if 'recommendations' not in result:
            result['recommendations'] = []
        if 'executive_summary' not in result:
            result['executive_summary'] = 'Risk assessment completed'
        
        return result
    
    def format_comparisons(self, comparisons: List[Dict]) -> str:
        """Format comparisons for prompt"""
        formatted = []
        
        for i, comp in enumerate(comparisons[:10], 1):  # Max 10 for context
            clause = comp.get('clause', {})
            comparison = comp.get('comparison', {})
            
            formatted.append(f"""
{i}. Clause Type: {clause.get('type', 'unknown')}
   Risk Level: {clause.get('risk_level', 'unknown')}
   Favorability: {comparison.get('favorability_score', 0):.2f}
   Key Issues: {', '.join(comparison.get('risks', [])[:2])}
""")
        
        return "\n".join(formatted)
    
    def calculate_risk_level(self, high_risk_count: int, total_count: int) -> str:
        """Calculate overall risk level"""
        if total_count == 0:
            return 'unknown'
        
        high_risk_ratio = high_risk_count / total_count
        
        if high_risk_ratio >= 0.4:
            return 'high'
        elif high_risk_ratio >= 0.2:
            return 'medium'
        else:
            return 'low'
    
    def calculate_risk_score(self, high_risk_count: int, total_count: int, avg_favorability: float) -> float:
        """Calculate numerical risk score"""
        if total_count == 0:
            return 0.5
        
        # Combine high risk ratio and favorability
        high_risk_ratio = high_risk_count / total_count
        favorability_factor = (1 - avg_favorability) / 2  # Convert -1..1 to 0..1
        
        # Weighted average
        risk_score = (high_risk_ratio * 0.6) + (favorability_factor * 0.4)
        
        return min(max(risk_score, 0.0), 1.0)  # Clamp to 0-1
