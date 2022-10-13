.PHONY:gen
bench:gen
	go test . -bench=. -benchmem -v 
gen:
	go install github.com/milkbobo/fishgoweb/language/querygen
	rm -rf testdata/testdata_querygen.go
	querygen -r github.com/milkbobo/fishgoweb/language/querygen/testdata
test:gen
	go test . -v
