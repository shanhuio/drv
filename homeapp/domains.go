package homeapp

// DomainMap is the domain mapping for one app.
type DomainMap struct {
	App string                  `json:",omitempty"`
	Map map[string]*DomainEntry `json:",omitempty"`
}

// DomainEntry is an entry for an application domain map.
type DomainEntry struct {
	Dest string
}

// Domains is a table that saves the application domain mapping.
type Domains interface {
	// Set sets the domain mapping of an application.
	Set(m *DomainMap) error

	// Clear clears the domain mapping of an application.
	Clear(app string) error
}
