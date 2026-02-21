"""
Vector Service for semantic search using ChromaDB
"""
import chromadb
from chromadb.config import Settings
from sentence_transformers import SentenceTransformer
import uuid
from typing import Dict, List, Any, Optional


class VectorService:
    def __init__(self, persist_dir="./chroma_db"):
        """Initialize ChromaDB client and embedding model"""
        # Get or create collection
        print(f"[VectorService] Initializing ChromaDB at {persist_dir}...", flush=True)
        self.client = chromadb.PersistentClient(path=persist_dir)
        
        print(f"[VectorService] Accessing collection: contracts...", flush=True)
        self.collection = self.client.get_or_create_collection(
            name="contracts",
            metadata={"hnsw:space": "cosine"}
        )
        
        # Use sentence-transformers for embeddings
        print("[VectorService] Loading SentenceTransformer model: all-MiniLM-L6-v2...", flush=True)
        self.embedder = SentenceTransformer('all-MiniLM-L6-v2')
        print("[VectorService] Model loaded successfully.", flush=True)
    
    def index_contract(self, contract_id: str, contract_name: str, clauses: List[Dict]) -> Dict[str, Any]:
        """
        Index all clauses from a contract
        
        Args:
            contract_id: Unique contract identifier
            contract_name: Human-readable contract name
            clauses: List of clause dicts with 'text', 'type', 'risk_level'
        
        Returns:
            Dict with indexing status
        """
        if not clauses:
            return {
                'contract_id': contract_id,
                'clauses_indexed': 0,
                'status': 'error',
                'message': 'No clauses provided'
            }
        
        documents = []
        metadatas = []
        ids = []
        
        for i, clause in enumerate(clauses):
            clause_id = f"{contract_id}_clause_{i}"
            documents.append(clause['text'])
            metadatas.append({
                'contract_id': contract_id,
                'contract_name': contract_name,
                'clause_type': clause.get('type', 'unknown'),
                'risk_level': clause.get('risk_level', 'medium'),
                'position': i
            })
            ids.append(clause_id)
        
        try:
            self.collection.add(
                documents=documents,
                metadatas=metadatas,
                ids=ids
            )
            
            return {
                'contract_id': contract_id,
                'contract_name': contract_name,
                'clauses_indexed': len(clauses),
                'status': 'success'
            }
        except Exception as e:
            return {
                'contract_id': contract_id,
                'clauses_indexed': 0,
                'status': 'error',
                'message': str(e)
            }
    
    def search_similar_clauses(
        self, 
        query_text: str, 
        top_k: int = 5,
        filters: Optional[Dict] = None
    ) -> Dict[str, Any]:
        """
        Search for similar clauses using semantic search
        
        Args:
            query_text: Text to search for
            top_k: Number of results to return
            filters: Optional metadata filters (e.g., {'clause_type': 'indemnification'})
        
        Returns:
            Dict with query and results list
        """
        try:
            where_clause = filters if filters else None
            
            results = self.collection.query(
                query_texts=[query_text],
                n_results=top_k,
                where=where_clause
            )
            
            # Format results
            formatted_results = []
            if results['documents'] and results['documents'][0]:
                for doc, meta, dist in zip(
                    results['documents'][0],
                    results['metadatas'][0],
                    results['distances'][0]
                ):
                    formatted_results.append({
                        'text': doc,
                        'metadata': meta,
                        'distance': float(dist),
                        'similarity': float(1 - dist)  # Convert distance to similarity
                    })
            
            return {
                'query': query_text,
                'top_k': top_k,
                'results': formatted_results,
                'count': len(formatted_results)
            }
        except Exception as e:
            return {
                'query': query_text,
                'results': [],
                'count': 0,
                'error': str(e)
            }
    
    def get_contract_clauses(self, contract_id: str) -> Dict[str, Any]:
        """Get all clauses for a specific contract"""
        try:
            results = self.collection.get(
                where={"contract_id": contract_id}
            )
            
            clauses = []
            if results['ids']:
                for doc, meta, id in zip(
                    results['documents'],
                    results['metadatas'],
                    results['ids']
                ):
                    clauses.append({
                        'id': id,
                        'text': doc,
                        'metadata': meta
                    })
            
            return {
                'contract_id': contract_id,
                'clauses': clauses,
                'count': len(clauses)
            }
        except Exception as e:
            return {
                'contract_id': contract_id,
                'clauses': [],
                'count': 0,
                'error': str(e)
            }
    
    def delete_contract(self, contract_id: str) -> Dict[str, Any]:
        """Remove all clauses for a contract from the index"""
        try:
            results = self.collection.get(
                where={"contract_id": contract_id}
            )
            
            if results['ids']:
                self.collection.delete(ids=results['ids'])
                return {
                    'contract_id': contract_id,
                    'deleted_count': len(results['ids']),
                    'status': 'success'
                }
            else:
                return {
                    'contract_id': contract_id,
                    'deleted_count': 0,
                    'status': 'not_found'
                }
        except Exception as e:
            return {
                'contract_id': contract_id,
                'deleted_count': 0,
                'status': 'error',
                'message': str(e)
            }
    
    def get_stats(self) -> Dict[str, Any]:
        """Get collection statistics"""
        try:
            count = self.collection.count()
            return {
                'total_clauses': count,
                'collection_name': self.collection.name,
                'status': 'healthy'
            }
        except Exception as e:
            return {
                'status': 'error',
                'message': str(e)
            }
