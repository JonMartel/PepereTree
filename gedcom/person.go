package gedcom

import (
	"fmt"
	"time"
)

//Person represents an individual as defined by the gedcom file format
type Person struct {
	ID          uint64
	Fullname    string
	Title       string
	Gender      string
	Occupation  string
	Aliases     []string
	Notes       []string
	Events      map[string]*event
	HeadFamIDs  []uint64
	ChildFamIDs []uint64
	Media       []string
	LastUpdate  time.Time
}

func (p *Person) String() string {
	if p.Events["BIRT"] != nil {
		birthEvent := p.Events["BIRT"]
		return fmt.Sprintf("%s|%s|%d|%d|%d|%v|%v", p.Fullname, p.Gender, birthEvent.EventYear,
			birthEvent.EventMonth, birthEvent.EventDay, p.HeadFamIDs, p.ChildFamIDs)
	}

	return fmt.Sprintf("%s|%s|%v|%v", p.Fullname, p.Gender, p.HeadFamIDs, p.ChildFamIDs)
}
