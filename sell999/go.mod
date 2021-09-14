module sell999

go 1.16

replace github.com/milkbobo/fishgoweb => ./extern/fishgoweb

require (
	github.com/milkbobo/fishgoweb v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.7.2
)
