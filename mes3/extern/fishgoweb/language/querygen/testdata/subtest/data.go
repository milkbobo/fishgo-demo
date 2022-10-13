package subtest

import (
	. "github.com/milkbobo/fishgoweb/language"
)

type Address struct {
	AddressId int
	City      string
}

func logic() {
	QueryColumn([]Address{}, "City")
}
