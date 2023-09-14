package persist

type (

	// Setter represents a set interface
	Setter interface {
		Set(interface{}) error
	}

	// Getter represents a get interface
	Getter interface {
		Get(interface{}) error
	}

	// Persist is the interface that groups the basic Marshal and Unmarshal methods.
	Persist interface {
		Setter
		Getter
	}
)
