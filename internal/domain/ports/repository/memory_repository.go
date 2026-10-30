package repositories

type IMemoryRepository interface {
	Save(id string, content string, category string) error
	RetrieveRelevant(query string) string
}
