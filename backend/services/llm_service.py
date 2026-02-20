"""
LLM Service for contract analysis using Groq API or Ollama
"""
import requests
import json
import os
from typing import Dict, Any, Optional


class LLMService:
    def __init__(self, provider: str = "groq", base_url: str = "http://localhost:11434"):
        """
        Initialize LLM client
        
        Args:
            provider: "groq" (cloud API) or "ollama" (local)
            base_url: Ollama base URL (only used if provider="ollama")
        """
        self.provider = provider
        self.base_url = base_url
        self.groq_api_key = os.getenv("GROQ_API_KEY", "")
        
        # Default models
        if provider == "groq":
            self.model_base = "llama-3.2-3b-preview"  # Groq model name
            self.model_finetuned = "llama-3.2-3b-preview"  # Same for now, will use system prompt
        else:
            self.model_base = "llama3.2:3b"
            self.model_finetuned = "legal-contract-analyzer"
    
    def complete(self, prompt: str, model: Optional[str] = None, 
                 use_finetuned: bool = True, temperature: float = 0.7, 
                 max_tokens: int = 2000) -> str:
        """
        Generate completion using Groq or Ollama
        
        Args:
            prompt: Input prompt
            model: Model name override (optional)
            use_finetuned: Use fine-tuned model (True) or base model (False)
            temperature: Sampling temperature
            max_tokens: Maximum tokens to generate
        
        Returns:
            Generated text
        """
        # Select model
        if model:
            model_name = model
        else:
            model_name = self.model_finetuned if use_finetuned else self.model_base
        
        # Add legal system prompt if using fine-tuned
        if use_finetuned and self.provider == "groq":
            system_prompt = """You are a legal contract analysis expert specialized in identifying clauses, risks, and obligations in legal documents. Analyze contracts carefully and provide detailed, accurate information."""
            full_prompt = f"{system_prompt}\n\n{prompt}"
        else:
            full_prompt = prompt
        
        try:
            if self.provider == "groq":
                return self._complete_groq(model_name, full_prompt, temperature, max_tokens)
            else:
                return self._complete_ollama(model_name, full_prompt, temperature, max_tokens)
        
        except Exception as e:
            return f"Error: {str(e)}"
    
    def _complete_groq(self, model: str, prompt: str, temperature: float, max_tokens: int) -> str:
        """Complete using Groq API"""
        if not self.groq_api_key:
            return "Error: GROQ_API_KEY not set in environment"
        
        try:
            response = requests.post(
                "https://api.groq.com/openai/v1/chat/completions",
                headers={
                    "Authorization": f"Bearer {self.groq_api_key}",
                    "Content-Type": "application/json"
                },
                json={
                    "model": model,
                    "messages": [
                        {"role": "user", "content": prompt}
                    ],
                    "temperature": temperature,
                    "max_tokens": max_tokens
                },
                timeout=120
            )
            
            if response.status_code == 200:
                result = response.json()
                return result['choices'][0]['message']['content']
            else:
                return f"Groq API Error: {response.status_code} - {response.text}"
        
        except requests.exceptions.RequestException as e:
            return f"Error connecting to Groq: {str(e)}"
    
    def _complete_ollama(self, model: str, prompt: str, temperature: float, max_tokens: int) -> str:
        """Complete using Ollama"""
        try:
            response = requests.post(
                f"{self.base_url}/api/generate",
                json={
                    "model": model,
                    "prompt": prompt,
                    "stream": False,
                    "options": {
                        "temperature": temperature,
                        "num_predict": max_tokens
                    }
                },
                timeout=120
            )
            
            if response.status_code == 200:
                result = response.json()
                return result.get('response', '')
            else:
                return f"Ollama Error: {response.status_code} - {response.text}"
        
        except requests.exceptions.RequestException as e:
            return f"Error connecting to Ollama: {str(e)}"
    
    def extract_json(self, text: str) -> Optional[Dict]:
        """Extract JSON from LLM response"""
        try:
            # Try to find JSON in the response
            start = text.find('{')
            end = text.rfind('}') + 1
            
            if start != -1 and end > start:
                json_str = text[start:end]
                return json.loads(json_str)
            
            # If no JSON found, try parsing the whole text
            return json.loads(text)
        
        except json.JSONDecodeError:
            return None
    
    def is_available(self) -> bool:
        """Check if LLM service is available"""
        if self.provider == "groq":
            return bool(self.groq_api_key)
        else:
            try:
                response = requests.get(f"{self.base_url}/api/tags", timeout=5)
                return response.status_code == 200
            except:
                return False
    
    def list_models(self) -> list:
        """List available models"""
        if self.provider == "groq":
            return [
                "llama-3.2-3b-preview",
                "llama-3.2-1b-preview", 
                "llama-3.1-8b-instant",
                "mixtral-8x7b-32768"
            ]
        else:
            try:
                response = requests.get(f"{self.base_url}/api/tags", timeout=5)
                if response.status_code == 200:
                    data = response.json()
                    return [model['name'] for model in data.get('models', [])]
                return []
            except:
                return []
