package gedcom

import (
	"fmt"
)

//Convenience map for all valid event types
var eventTypes = map[string]bool{
	"BIRT": true,
	"DEAT": true,
	"BAPM": true,
	"BURI": true,
}

//Event represents some event in a noun's history, per the gedcom format
type event struct {
	EventYear  int
	EventMonth int
	EventDay   int
	EventType  string
	SourceID   int
	Location   string
}

func (e *event) String() string {
	return fmt.Sprintf("Year: %d Month: %d Day: %d Location: %s Type: %s sourceid: %d",
		e.EventYear, e.EventMonth, e.EventDay, e.Location, e.EventType, e.SourceID)
}
