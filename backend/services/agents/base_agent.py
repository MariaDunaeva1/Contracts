"""
Base Agent class for contract analysis
"""
from abc import ABC, abstractmethod
from typing import Dict, Any


class BaseAgent(ABC):
    """Abstract base class for all agents"""
    
    def __init__(self, llm_service, use_finetuned: bool = True, model_name: str = None):
        """
        Initialize agent with LLM service
        
        Args:
            llm_service: LLMService instance for text generation
            use_finetuned: Use fine-tuned model (True) or base model (False)
            model_name: Specific model name override (optional)
        """
        self.llm = llm_service
        self.use_finetuned = use_finetuned
        self.model_name = model_name
    
    @abstractmethod
    def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Execute agent task
        
        Args:
            input_data: Input data for the agent
        
        Returns:
            Dict with agent results
        """
        pass
    
    def complete(self, prompt: str, **kwargs) -> str:
        """
        Wrapper for LLM completion with model selection
        
        Args:
            prompt: Input prompt
            **kwargs: Additional arguments for LLM
        
        Returns:
            Generated text
        """
        return self.llm.complete(prompt, model=self.model_name, use_finetuned=self.use_finetuned, **kwargs)
    
    def parse_response(self, response: str) -> Dict[str, Any]:
        """
        Parse LLM response to structured format
        
        Args:
            response: Raw LLM response
        
        Returns:
            Parsed dict or error dict
        """
        parsed = self.llm.extract_json(response)
        
        if parsed:
            return parsed
        else:
            return {
                'error': 'Failed to parse response',
                'raw_response': response
            }
