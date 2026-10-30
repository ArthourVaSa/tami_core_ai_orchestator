package repositories

import (
	"fmt"
)

// ChromaRepo es la implementación concreta de IMemoryRepository
type ChromaRepo struct {
	baseURL    string
	collection string
}

// NewChromaRepo inicializa la conexión a la base de datos vectorial
func NewChromaRepo(baseURL string, collection string) *ChromaRepo {
	return &ChromaRepo{
		baseURL:    baseURL,
		collection: collection,
	}
}

// Save persiste la memoria en el espacio vectorial
func (r *ChromaRepo) Save(id string, content string, category string) error {
	// TODO: Implementar petición HTTP POST a la API REST de ChromaDB
	fmt.Printf("💾 [CHROMA DB] Guardando memoria en colección '%s'...\n", r.collection)
	fmt.Printf("   -> ID: %s | Categoría: %s\n", id, category)
	fmt.Printf("   -> Contenido: %s\n", content)

	return nil
}

// RetrieveRelevant busca los embeddings más cercanos a la query
func (r *ChromaRepo) RetrieveRelevant(query string) string {
	// TODO: Implementar búsqueda vectorial HTTP a ChromaDB
	fmt.Printf("🔍 [CHROMA DB] Buscando contexto vectorial para: '%s'\n", query)

	// Por ahora devolvemos un contexto vacío para que Ollama responda con su conocimiento base
	return ""
}
