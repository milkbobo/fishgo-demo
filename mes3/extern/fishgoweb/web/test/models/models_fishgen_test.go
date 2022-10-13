package test

import (
	. "github.com/milkbobo/fishgoweb/web"
	"testing"
)

type testFishGenStruct struct{}

func TestTest(t *testing.T) {
	RunTest(t, &testFishGenStruct{})
}
