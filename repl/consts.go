package repl

const (
	// JSONArray means array struct [] in json
	JSONArray = iota
	// JSONMap means map struct {} in json
	JSONMap
	// JSONString means string "" in json
	JSONString
	// JSONInt means int in json
	JSONInt
	// JSONEnd means nothing valuable found
	JSONEnd
)

const (
	CommandSetTypeLocal   = "local"
	CommandSetTypeRemote  = "remote"
	CommandSetTypeAccount = "account"
)
