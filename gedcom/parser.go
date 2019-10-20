package gedcom

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

//Individuals : Map of all the individuals we've found by id
var individuals = make(map[uint64]*Person)

//Families : Families, by id
var families = make(map[uint64]*Family)

//Sources : All the backing source info we have for a person by id
var sources = make(map[uint64]*source)

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

//GetGedcomData returns all the parsed data to the callee
func GetGedcomData() (people map[uint64]*Person, families map[uint64]*Family, sources map[uint64]*source) {
	return individuals, families, sources
}

//Parse accepts a path to a gedcom file and parses out the data contained within
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
				individuals[person.ID] = person
			} else {
				fmt.Println("Error creating a person:", err)
			}
		case "FAM":
			fam, err := parseFamily(lines)
			if err == nil {
				families[fam.ID] = fam
			} else {
				fmt.Println("Error creating family:", err)
			}
		case "SOUR":
			sour, err := parseSource(lines)
			if err == nil {
				//Update the source with the abbr and note fields
				//If it isnt in the map, nothing references it
				oldSour := sources[sour.ID]
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
	dude.Events = make(map[string]*event)

	id, err := parseID(lines[0])

	if err == nil {
		dude.ID = id
		//Next, lets go ahead and parse the remaining lines
		var ev *event = nil
		var src *source = nil
		for _, line := range lines[1:] {
			lineComponents := strings.SplitN(line, " ", 3)
			level, _ := strconv.ParseInt(lineComponents[0], 10, 64)
			linetype := lineComponents[1]
			if level == 1 {
				ev = nil
				_, isEvent := eventTypes[linetype]
				if isEvent {
					ev = new(event)
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
						famID, err := parseComponentID(lineComponents[2])
						if err == nil {
							if dude.HeadFamIDs != nil {
								dude.HeadFamIDs = append(dude.HeadFamIDs, famID)
							} else {
								dude.HeadFamIDs = make([]uint64, 1)
								dude.HeadFamIDs[0] = famID
							}
						}
					case "FAMC":
						famid, err := parseComponentID(lineComponents[2])
						if err == nil {
							if dude.ChildFamIDs != nil {
								dude.ChildFamIDs = append(dude.ChildFamIDs, famid)
							} else {
								dude.ChildFamIDs = make([]uint64, 1)
								dude.ChildFamIDs[0] = famid
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
					sid, err := strconv.ParseUint(idPortion, 10, 64)
					if err == nil {
						src = new(source)
						src.ID = sid
						sources[src.ID] = src
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
						quality, err := strconv.ParseUint(lineComponents[2], 10, 64)
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
	id, err := parseID(lines[0])

	if err == nil {
		fam.ID = id

		for _, line := range lines[1:] {
			lineComponents := strings.SplitN(line, " ", 3)
			//level, _ := strconv.ParseInt(lineComponents[0], 10, 64)
			linetype := lineComponents[1]
			switch linetype {
			//1 HUSB @I5809@
			//1 WIFE @I4517@
			//1 CHIL @I5810@
			case "HUSB":
				idString := lineComponents[2]
				parsedID, err := strconv.ParseUint(idString[2:len(idString)-1], 10, 64)
				if err == nil {
					fam.Father = parsedID
				}
			case "WIFE":
				idString := lineComponents[2]
				parsedID, err := strconv.ParseUint(idString[2:len(idString)-1], 10, 64)
				if err == nil {
					fam.Mother = parsedID
				}
			case "CHIL":
				idString := lineComponents[2]
				parsedID, err := strconv.ParseUint(idString[2:len(idString)-1], 10, 64)
				if err == nil {
					if fam.ChildIDs == nil {
						fam.ChildIDs = make([]uint64, 1)
						fam.ChildIDs[0] = parsedID
					} else {
						fam.ChildIDs = append(fam.ChildIDs, parsedID)
					}
				}

			}

		}
	}
	return fam, nil
}

func parseSource(lines []string) (*source, error) {
	sour := new(source)
	id, err := parseID(lines[0])
	if err == nil {
		sour.ID = id

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

func parseID(line string) (uint64, error) {
	//First line should be like one of these:
	//0 @S1234@ SOUR
	//0 @F1234@ FAM
	//0 @I1234@ INDI
	indiComponents := strings.SplitN(line, " ", 3)

	//Take the id string, and get to the 'meat'
	idString := indiComponents[1]
	strID := idString[2 : len(idString)-1]
	pID, err := strconv.ParseUint(strID, 10, 64)

	return pID, err
}

func parseComponentID(component string) (uint64, error) {
	pID, err := strconv.ParseUint(component[2:len(component)-1], 10, 64)
	return pID, err
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
}

func parseFileString(line string) string {
	//C:\BK6RAC\Picture\6032A.JPG
	//Remove everything before the 4th index splitting on '\'

	return line
}
