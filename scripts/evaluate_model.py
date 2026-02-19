#!/usr/bin/env python3
"""
Model Evaluation Script
Compares base model vs fine-tuned model performance
"""

import json
import time
import sys
import argparse
from typing import List, Dict, Any
from pathlib import Path

try:
    import requests
except ImportError:
    print("Installing required packages...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "requests"])
    import requests

try:
    from sklearn.metrics import accuracy_score, f1_score, precision_score, recall_score
except ImportError:
    print("Installing scikit-learn...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "scikit-learn"])
    from sklearn.metrics import accuracy_score, f1_score, precision_score, recall_score


class OllamaClient:
    """Simple Ollama API client"""
    
    def __init__(self, base_url: str = "http://localhost:11434"):
        self.base_url = base_url
    
    def generate(self, model: str, prompt: str, temperature: float = 0.1) -> Dict[str, Any]:
        """Generate text using Ollama"""
        url = f"{self.base_url}/api/generate"
        payload = {
            "model": model,
            "prompt": prompt,
            "temperature": temperature,
            "stream": False
        }
        
        try:
            response = requests.post(url, json=payload, timeout=60)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            print(f"Error calling Ollama: {e}")
            return {"response": ""}
    
    def list_models(self) -> List[str]:
        """List available models"""
        url = f"{self.base_url}/api/tags"
        try:
            response = requests.get(url, timeout=10)
            response.raise_for_status()
            data = response.json()
            return [model["name"] for model in data.get("models", [])]
        except Exception as e:
            print(f"Error listing models: {e}")
            return []


def load_test_set(test_set_path: str) -> List[Dict[str, Any]]:
    """Load test dataset"""
    with open(test_set_path, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    # Handle different formats
    if isinstance(data, list):
        return data
    elif isinstance(data, dict) and 'data' in data:
        return data['data']
    else:
        raise ValueError("Invalid test set format")


def extract_label(response: str, labels: List[str]) -> str:
    """Extract label from model response"""
    response_lower = response.lower().strip()
    
    # Try exact match
    for label in labels:
        if label.lower() in response_lower:
            return label
    
    # Try first word
    first_word = response_lower.split()[0] if response_lower else ""
    for label in labels:
        if first_word == label.lower():
            return label
    
    # Default to first label
    return labels[0] if labels else "unknown"


def evaluate_model(
    ollama: OllamaClient,
    model_name: str,
    test_set: List[Dict[str, Any]],
    labels: List[str]
) -> Dict[str, Any]:
    """Evaluate a model on test set"""
    
    predictions = []
    ground_truth = []
    response_times = []
    examples = []
    
    print(f"\nEvaluating {model_name}...")
    
    for i, example in enumerate(test_set):
        if i % 10 == 0:
            print(f"  Progress: {i}/{len(test_set)}")
        
        prompt = example.get('prompt', example.get('text', ''))
        expected = example.get('label', example.get('expected', ''))
        
        # Generate prediction
        start = time.time()
        result = ollama.generate(model_name, prompt)
        elapsed = time.time() - start
        
        response = result.get('response', '').strip()
        predicted = extract_label(response, labels)
        
        predictions.append(predicted)
        ground_truth.append(expected)
        response_times.append(elapsed * 1000)  # Convert to ms
        
        # Store example for comparison
        if i < 20:  # Store first 20 examples
            examples.append({
                "input": prompt[:100],  # Truncate long inputs
                "expected": expected,
                "predicted": predicted,
                "correct": predicted == expected,
                "response": response[:200],  # Truncate long responses
                "response_time_ms": round(elapsed * 1000, 2)
            })
    
    # Calculate metrics
    try:
        accuracy = accuracy_score(ground_truth, predictions)
        precision = precision_score(ground_truth, predictions, average='weighted', zero_division=0)
        recall = recall_score(ground_truth, predictions, average='weighted', zero_division=0)
        f1 = f1_score(ground_truth, predictions, average='weighted', zero_division=0)
    except Exception as e:
        print(f"Warning: Error calculating metrics: {e}")
        accuracy = precision = recall = f1 = 0.0
    
    avg_response_time = sum(response_times) / len(response_times) if response_times else 0
    
    return {
        "accuracy": round(accuracy, 4),
        "precision": round(precision, 4),
        "recall": round(recall, 4),
        "f1_score": round(f1, 4),
        "avg_response_time_ms": round(avg_response_time, 2),
        "total_samples": len(test_set),
        "examples": examples
    }


def calculate_improvement(base_results: Dict, finetuned_results: Dict) -> Dict[str, str]:
    """Calculate improvement metrics"""
    
    def calc_delta(base_val: float, ft_val: float) -> str:
        if base_val == 0:
            return "N/A"
        delta = ((ft_val - base_val) / base_val) * 100
        sign = "+" if delta >= 0 else ""
        return f"{sign}{delta:.1f}%"
    
    return {
        "accuracy_delta": calc_delta(base_results["accuracy"], finetuned_results["accuracy"]),
        "f1_score_delta": calc_delta(base_results["f1_score"], finetuned_results["f1_score"]),
        "precision_delta": calc_delta(base_results["precision"], finetuned_results["precision"]),
        "recall_delta": calc_delta(base_results["recall"], finetuned_results["recall"]),
        "response_time_delta": calc_delta(
            base_results["avg_response_time_ms"],
            finetuned_results["avg_response_time_ms"]
        )
    }


def main():
    parser = argparse.ArgumentParser(description="Evaluate model performance")
    parser.add_argument("--test-set", required=True, help="Path to test set JSON file")
    parser.add_argument("--base-model", default="llama3.2:3b", help="Base model name")
    parser.add_argument("--finetuned-model", required=True, help="Fine-tuned model name")
    parser.add_argument("--labels", nargs="+", default=["positive", "negative", "neutral"],
                       help="List of possible labels")
    parser.add_argument("--output", default="evaluation_results.json", help="Output file")
    parser.add_argument("--ollama-url", default="http://localhost:11434", help="Ollama API URL")
    
    args = parser.parse_args()
    
    # Initialize Ollama client
    ollama = OllamaClient(args.ollama_url)
    
    # Check available models
    print("Checking available models...")
    available_models = ollama.list_models()
    print(f"Available models: {available_models}")
    
    if args.base_model not in available_models:
        print(f"Warning: Base model '{args.base_model}' not found in Ollama")
    
    if args.finetuned_model not in available_models:
        print(f"Warning: Fine-tuned model '{args.finetuned_model}' not found in Ollama")
    
    # Load test set
    print(f"\nLoading test set from {args.test_set}...")
    test_set = load_test_set(args.test_set)
    print(f"Loaded {len(test_set)} test examples")
    
    # Evaluate base model
    base_results = evaluate_model(ollama, args.base_model, test_set, args.labels)
    
    # Evaluate fine-tuned model
    finetuned_results = evaluate_model(ollama, args.finetuned_model, test_set, args.labels)
    
    # Calculate improvement
    improvement = calculate_improvement(base_results, finetuned_results)
    
    # Combine examples for side-by-side comparison
    combined_examples = []
    for i in range(min(len(base_results["examples"]), len(finetuned_results["examples"]))):
        base_ex = base_results["examples"][i]
        ft_ex = finetuned_results["examples"][i]
        
        combined_examples.append({
            "input": base_ex["input"],
            "expected": base_ex["expected"],
            "base_model_output": base_ex["predicted"],
            "base_model_correct": base_ex["correct"],
            "fine_tuned_output": ft_ex["predicted"],
            "fine_tuned_correct": ft_ex["correct"],
            "winner": "fine_tuned" if ft_ex["correct"] and not base_ex["correct"] else
                     "base" if base_ex["correct"] and not ft_ex["correct"] else
                     "tie"
        })
    
    # Create final results
    results = {
        "base_model": {
            "name": args.base_model,
            "metrics": base_results
        },
        "fine_tuned": {
            "name": args.finetuned_model,
            "metrics": finetuned_results
        },
        "improvement": improvement,
        "examples": combined_examples,
        "test_set_size": len(test_set),
        "labels": args.labels,
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
    }
    
    # Save results
    output_path = Path(args.output)
    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(results, f, indent=2)
    
    print(f"\n{'='*60}")
    print("EVALUATION RESULTS")
    print(f"{'='*60}")
    print(f"\nBase Model ({args.base_model}):")
    print(f"  Accuracy:  {base_results['accuracy']:.2%}")
    print(f"  F1 Score:  {base_results['f1_score']:.4f}")
    print(f"  Precision: {base_results['precision']:.4f}")
    print(f"  Recall:    {base_results['recall']:.4f}")
    print(f"  Avg Time:  {base_results['avg_response_time_ms']:.2f}ms")
    
    print(f"\nFine-tuned Model ({args.finetuned_model}):")
    print(f"  Accuracy:  {finetuned_results['accuracy']:.2%}")
    print(f"  F1 Score:  {finetuned_results['f1_score']:.4f}")
    print(f"  Precision: {finetuned_results['precision']:.4f}")
    print(f"  Recall:    {finetuned_results['recall']:.4f}")
    print(f"  Avg Time:  {finetuned_results['avg_response_time_ms']:.2f}ms")
    
    print(f"\nImprovement:")
    print(f"  Accuracy:  {improvement['accuracy_delta']}")
    print(f"  F1 Score:  {improvement['f1_score_delta']}")
    print(f"  Precision: {improvement['precision_delta']}")
    print(f"  Recall:    {improvement['recall_delta']}")
    
    print(f"\nResults saved to: {output_path}")
    print(f"{'='*60}\n")


if __name__ == "__main__":
    main()
