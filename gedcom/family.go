package gedcom

import "fmt"

//Family represents a family unit as defined by the gedcom format
//Thus, it has a Father and Mother and a number of children
type Family struct {
	ID       int
	Father   int
	Mother   int
	ChildIDs []int
	Notes    []string
}

func (f *Family) String() string {
	return fmt.Sprintf("%d|%d|%d", f.ID, f.Father, f.Mother)
}

//DisplayFamily prints out to stdout the info for the specified family id
func DisplayFamily(id int64) {
	family := Families[int(id)]

	if family != nil {

		fmt.Println(family)
		fmt.Println(Individuals[family.Father])
		fmt.Println(Individuals[family.Mother])
		for _, child := range family.ChildIDs {
			fmt.Println(Individuals[child])

			person := Individuals[child]
			for _, ev := range person.Events {
				fmt.Println(ev)
			}
		}
	} else {
		fmt.Printf("No family found for id %d\n", id)
	}
}
