package gedcom

//Source represents where a particular piece of information can be sourced
type source struct {
	ID      uint64
	Quality int
	Page    string
	Note    string
	Abbr    string
}
