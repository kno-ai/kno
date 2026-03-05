package skills

// Store provides access to skill/policy markdown documents.
type Store interface {
	Get(name string) (string, error)
	List() ([]string, error)
}
