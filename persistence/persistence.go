package persistence

type (

	// Setter represents a set interface
	Setter interface {
		Set(interface{}) error
	}

	// Getter represents a get interface
	Getter interface {
		Get(interface{}) error
	}

	// Persistence is the interface that groups the basic Marshal and Unmarshal methods.
	Persistence interface {
		Setter
		Getter
	}
)
