import chromadb
from app.domain.ports.memory_repository import IMemoryRepository

class ChromaRepository(IMemoryRepository):
    def __init__(self, persist_directory: str = "./chroma_db"):
        # Usamos PersistentClient para que los datos sobrevivan a los reinicios
        self.client = chromadb.PersistentClient(path=persist_directory)
        self.collection = self.client.get_or_create_collection(name="project_memory")

    def save(self, memory_id: str, content: str, category: str) -> None:
        self.collection.add(
            documents=[content],
            metadatas=[{"category": category}],
            ids=[memory_id]
        )

    def retrieve_relevant(self, query: str, limit: int = 2) -> str:
        results = self.collection.query(
            query_texts=[query],
            n_results=limit
        )
        
        # Si no hay resultados o la base de datos está vacía
        if not results['documents'] or not results['documents'][0]:
            return ""
        
        # Unimos los documentos encontrados separándolos por un salto de línea
        return "\n".join(results['documents'][0])