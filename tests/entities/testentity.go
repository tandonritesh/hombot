package main

import (
	"hombot/errors"
	"hombot/intents/entityset"
	"log"
)

type Test struct {
	val1 int
	val2 string
}

func (t *Test) hash() int {
	return 1
}

func main() {
	var entitySet entityset.EntitySet

	// m := make(map[Test]int)

	// m[Test{1, "A"}] = 1

	// x := m[Test{1, ""}]

	// log.Print(x)
	entitySet.Init()

	//var refString string = ("please turn on the fan at speed 4")
	var refString string = ("please turn on the fan")
	// var refString string = ("fan turn on")
	// var refString string = ("turn fan on")
	// var refString string = (" fan turn on")
	//var refString string = ("please turn on the fan ")
	// var refString string = ("please turn on the fan1")
	// var refString string = ("please turn on the 1fan")
	// var refString string = ("fan1 turn on")
	// var refString string = ("1fan turn on")
	// var refString string = ("turn fan1 on")

	entity, errCode := entitySet.SearchEntity(refString)
	if errCode != errors.SUCCESS {
		log.Fatalf("Search entity failed with %d", errCode)
	}

	values, maps := entitySet.GetEntity(entity)
	log.Printf("Length %d", len(*values))
	for i, val := range *values {
		log.Printf("Value is pos:%d: val: %v", i, val)
	}

	//var valueSet *entityset.ValueSet
	valueSet, errCode := entitySet.SearchValues(refString, values, maps)
	if errCode != errors.SUCCESS {
		log.Fatalf("Failed to get values")
	}
	log.Printf("Valueset = %d", valueSet)
}
