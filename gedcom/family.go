package gedcom

import "fmt"

//Family represents a family unit as defined by the gedcom format
//Thus, it has a Father and Mother and a number of children
type Family struct {
	ID       uint64
	Father   uint64
	Mother   uint64
	ChildIDs []uint64
	Notes    []string
}

func (f *Family) String() string {
	return fmt.Sprintf("%d|%d|%d", f.ID, f.Father, f.Mother)
}

//DisplayFamily prints out to stdout the info for the specified family id
func DisplayFamily(id uint64) {
	family := families[id]

	if family != nil {

		fmt.Println(family)
		fmt.Println(individuals[family.Father])
		fmt.Println(individuals[family.Mother])
		for _, child := range family.ChildIDs {
			fmt.Println(individuals[child])

			person := individuals[child]
			for _, ev := range person.Events {
				fmt.Println(ev)
			}
		}
	} else {
		fmt.Printf("No family found for id %d\n", id)
	}
}
