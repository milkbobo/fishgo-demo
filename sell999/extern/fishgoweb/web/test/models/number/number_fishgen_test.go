package number

import (
	. "github.com/milkbobo/fishgoweb/web"
	"testing"
)

type testFishGenStruct struct{}

func TestNumber(t *testing.T) {
	RunTest(t, &testFishGenStruct{})
}
