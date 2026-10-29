from abc import ABC, abstractmethod

class IMemoryRepository(ABC):
    """Abstract interface for memory repositories."""
    
    @abstractmethod
    def save(self, memory_id: str, content: str, category: str) -> None:
        pass

    @abstractmethod
    def retrieve_relevant(self, query: str, limit: int = 2) -> str:
        pass