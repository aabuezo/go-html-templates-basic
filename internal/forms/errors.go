package forms

type ValidationErrors map[string][]string

// Add adds an error message for a given form field
func (e ValidationErrors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get returns the first error message
func (e ValidationErrors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}

	return es[0]
}
