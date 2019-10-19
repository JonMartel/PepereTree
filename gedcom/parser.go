package gedcom

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

//Individuals : Map of all the individuals we've found by id
var Individuals = make(map[int]*Person)

//Families : Families, by id
var Families = make(map[int]*Family)

//Sources : All the backing source info we have for a person by id
var Sources = make(map[int]*Source)

//Convenience map to convert the dates in gedcom to usable ints
var months = map[string]int{
	"JAN": 1,
	"FEB": 2,
	"MAR": 3,
	"APR": 4,
	"MAY": 5,
	"JUN": 6,
	"JUL": 7,
	"AUG": 8,
	"SEP": 9,
	"OCT": 10,
	"NOV": 11,
	"DEC": 12,
}

//Convenience map for all valid event types
var eventtypes = map[string]bool{
	"BIRT": true,
	"DEAT": true,
	"BAPM": true,
	"BURI": true,
}

type Person struct {
	Id          int
	Fullname    string
	Title       string
	Gender      string
	Occupation  string
	Aliases     []string
	Notes       []string
	Events      map[string]*Event
	HeadFamIds  []int
	ChildFamIds []int
	Media       []string
	LastUpdate  time.Time
}

func (p *Person) String() string {
	if p.Events["BIRT"] != nil {
		birthEvent := p.Events["BIRT"]
		return fmt.Sprintf("%s|%s|%d|%s|%s|%s|%s", p.Fullname, p.Gender, birthEvent.EventYear, birthEvent.EventMonth, birthEvent.EventDay, p.HeadFamIds, p.ChildFamIds)
	}

	return fmt.Sprintf("%s|%s|%s|%s", p.Fullname, p.Gender, p.HeadFamIds, p.ChildFamIds)
}

type Family struct {
	Id       int
	Father   int
	Mother   int
	ChildIds []int
	Notes    []string
}

func (f *Family) String() string {
	return fmt.Sprintf("%d|%d|%d", f.Id, f.Father, f.Mother)
}

type Event struct {
	EventYear  int
	EventMonth int
	EventDay   int
	EventType  string
	SourceId   int
	Location   string
}

func (e *Event) String() string {
	return fmt.Sprintf("Year: %s Month: %s Day: %s Location: %s Type: %s sourceid: %d", e.EventYear, e.EventMonth, e.EventDay, e.Location, e.EventType, e.SourceId)
}

type Source struct {
	Id      int
	Quality int
	Page    string
	Note    string
	Abbr    string
}

type object struct {
	id   int
	name string
}

func Parse(filepath string) {
	fmt.Println("Gedcom Parser: ", filepath)

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var (
		objectLines []string = make([]string, 0)
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linetext := scanner.Text()
		split := strings.SplitN(linetext, " ", 2)
		//index 0 has the 'value' (0,1,2,3,4,5) for the field. 0 is the start of a record
		//index 1 has the field name (FAMC, NOTE, NAME, SEX, BIRT, etc
		//index 2 has the remainder, which will vary with field type

		lineType, err := strconv.ParseUint(split[0], 10, 8)
		if err == nil {
			switch lineType {
			case 0:
				//Did we have something previously? Let's parse them now
				if len(objectLines) > 0 {
					parseRecord(objectLines)
				}

				//Now, reset the slice and fill in the first element!
				objectLines = make([]string, 0)
				objectLines = append(objectLines, linetext)

			default:
				objectLines = append(objectLines, linetext)
			}
		}
	}

	if len(objectLines) > 0 {
		parseRecord(objectLines)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func parseRecord(lines []string) {
	components := strings.SplitN(lines[0], " ", 3)

	switch len(components) {
	case 3:
		switch components[2] {
		case "INDI":
			person, err := parsePerson(lines)
			if err == nil {
				Individuals[person.Id] = person
			} else {
				fmt.Println("Error creating a person:", err)
			}
		case "FAM":
			fam, err := parseFamily(lines)
			if err == nil {
				Families[fam.Id] = fam
			} else {
				fmt.Println("Error creating family:", err)
			}
		case "SOUR":
			sour, err := parseSource(lines)
			if err == nil {
				//Update the source with the abbr and note fields
				//If it isnt in the map, nothing references it
				oldSour := Sources[sour.Id]
				if oldSour != nil {
					oldSour.Abbr = sour.Abbr
					oldSour.Note = sour.Note
				}
			} else {
				fmt.Println("Error filling in the source", err)
			}
		}
	}
}

func parsePerson(lines []string) (*Person, error) {
	//First line should be like:
	//0 @I1234@ INDI
	dude := new(Person)
	dude.Events = make(map[string]*Event)

	id, err := parseId(lines[0])

	if err == nil {
		dude.Id = int(id)
		//Next, lets go ahead and parse the remaining lines
		var ev *Event = nil
		var src *Source = nil
		for _, line := range lines[1:] {
			lineComponents := strings.SplitN(line, " ", 3)
			level, _ := strconv.ParseInt(lineComponents[0], 10, 32)
			linetype := lineComponents[1]
			if level == 1 {
				ev = nil
				_, isEvent := eventtypes[linetype]
				if isEvent {
					ev = new(Event)
					dude.Events[linetype] = ev
					ev.EventType = linetype
				} else {

					switch linetype {
					case "NAME":
						dude.Fullname = lineComponents[2]
					case "SEX":
						dude.Gender = lineComponents[2]
					case "NOTE":
						dude.Notes = make([]string, 0)
					case "OCCU":
						dude.Occupation = lineComponents[2]
					case "FAMS":
						famid, err := parseComponentId(lineComponents[2])
						if err == nil {
							if dude.HeadFamIds != nil {
								dude.HeadFamIds = append(dude.HeadFamIds, int(famid))
							} else {
								dude.HeadFamIds = make([]int, 1)
								dude.HeadFamIds[0] = int(famid)
							}
						}
					case "FAMC":
						famid, err := parseComponentId(lineComponents[2])
						if err == nil {
							if dude.ChildFamIds != nil {
								dude.ChildFamIds = append(dude.ChildFamIds, int(famid))
							} else {
								dude.ChildFamIds = make([]int, 1)
								dude.ChildFamIds[0] = int(famid)
							}
						}
					}
				}
			} else if level == 2 {
				//Extends the current 'event'
				switch linetype {
				case "DATE":
					if ev != nil {
						dateComponents := strings.Split(lineComponents[2], " ")
						ev.EventYear, ev.EventMonth, ev.EventDay = parseDateString(dateComponents)
					}
				case "PLAC":
					if ev != nil {
						ev.Location = lineComponents[2]
					}
				case "SOUR":
					idPortion := lineComponents[2][1 : len(lineComponents[2])-1]
					sid, err := strconv.ParseUint(idPortion, 10, 32)
					if err == nil {
						src = new(Source)
						src.Id = int(sid)
						Sources[src.Id] = src
					}
				case "CONT":
					dude.Notes = append(dude.Notes, lineComponents[2])
				case "FILE":
					lowercase := strings.ToLower(lineComponents[2])

					if dude.Media == nil {
						dude.Media = make([]string, 1)
						dude.Media[0] = parseFileString(lowercase)
					} else {
						dude.Media = append(dude.Media, parseFileString(lowercase))
					}
				}
			} else if level == 3 {
				//Extends the source
				if src != nil {
					switch linetype {
					case "PAGE":
						src.Page = lineComponents[2]
					case "QUAY":
						quality, err := strconv.ParseUint(lineComponents[2], 10, 32)
						if err == nil {
							src.Quality = int(quality)
						}
					}
				}
			}
		}

	} else {
		dude = nil
		fmt.Println("Creation of person failed", err)
	}

	return dude, err
}

func parseFamily(lines []string) (*Family, error) {

	fam := new(Family)
	id, err := parseId(lines[0])

	if err == nil {
		fam.Id = int(id)

		for _, line := range lines[1:] {
			lineComponents := strings.SplitN(line, " ", 3)
			//level, _ := strconv.ParseInt(lineComponents[0], 10, 32)
			linetype := lineComponents[1]
			switch linetype {
			//1 HUSB @I5809@
			//1 WIFE @I4517@
			//1 CHIL @I5810@
			case "HUSB":
				idString := lineComponents[2]
				parsedId, err := strconv.ParseInt(idString[2:len(idString)-1], 10, 32)
				if err == nil {
					fam.Father = int(parsedId)
				}
			case "WIFE":
				idString := lineComponents[2]
				parsedId, err := strconv.ParseInt(idString[2:len(idString)-1], 10, 32)
				if err == nil {
					fam.Mother = int(parsedId)
				}
			case "CHIL":
				idString := lineComponents[2]
				parsedId, err := strconv.ParseInt(idString[2:len(idString)-1], 10, 32)
				if err == nil {
					if fam.ChildIds == nil {
						fam.ChildIds = make([]int, 1)
						fam.ChildIds[0] = int(parsedId)
					} else {
						fam.ChildIds = append(fam.ChildIds, int(parsedId))
					}
				}

			}

		}
	}
	return fam, nil
}

func parseSource(lines []string) (*Source, error) {
	sour := new(Source)
	id, err := parseId(lines[0])
	if err == nil {
		sour.Id = id

		for _, line := range lines[1:] {
			lineComponents := strings.SplitN(line, " ", 3)
			linetype := lineComponents[1]
			switch linetype {
			case "ABBR":
				sour.Abbr = lineComponents[2]
			case "NOTE":
				sour.Note = lineComponents[2]
			}
		}
	}

	return sour, err
}

func parseId(line string) (int, error) {
	//First line should be like one of these:
	//0 @S1234@ SOUR
	//0 @F1234@ FAM
	//0 @I1234@ INDI
	indiComponents := strings.SplitN(line, " ", 3)

	//Take the id string, and get to the 'meat'
	idString := indiComponents[1]
	strId := idString[2 : len(idString)-1]
	pId, err := strconv.ParseUint(strId, 10, 32)

	return int(pId), err
}

func parseComponentId(component string) (int, error) {
	pId, err := strconv.ParseInt(component[2:len(component)-1], 10, 32)
	return int(pId), err
}

//Returns year, month, day values for this date string
//Some dates will only have certain components, in which case
//nil is returned for the unspecified parts
func parseDateString(datecomps []string) (int, int, int) {

	switch len(datecomps) {
	case 1:
		//Just the year
		year, _ := strconv.Atoi(datecomps[0])
		return year, -1, -1
	case 2:
		//ABT YYYY
		year, _ := strconv.Atoi(datecomps[1])
		return year, -1, -1
	case 3:
		//11 DEC 1955
		day, _ := strconv.Atoi(datecomps[0])
		month := months[datecomps[1]]
		year, _ := strconv.Atoi(datecomps[2])
		return year, month, day
	default:
		return -1, -1, -1
	}

	return -1, -1, -1
}

func parseFileString(line string) string {
	//C:\BK6RAC\Picture\6032A.JPG
	//Remove everything before the 4th index splitting on '\'

	return line
}
