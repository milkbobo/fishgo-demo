package common

type CommonPage struct {
	PageSize  int
	PageIndex int
}

var CommonAllPage = CommonPage{PageSize: -1, PageIndex: -1}
