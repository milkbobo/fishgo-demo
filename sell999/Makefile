watch:
	-pkill -9 fishcmd
	-pkill -9 sell999
	-pkill -9 main
	nohup fishcmd run sell999 --watch &
	tail -f nohup.out
release:
	go run models/tool/startPrepare/startPrepare.go
	go generate
	go build
	sudo supervisorctl restart sell999
	sudo tail -f ../data/log/access.log
gen:
	go generate
test:
	go generate
	go test ./models/...
stop:
	sudo supervisorctl stop sell999
	-pkill -9 bee
	-pkill -9 goldenstatue
docker-build:
	sudo docker build . -t fildata
docker-run:
	# sudo docker run --network host -d -p 9000:9000 -p 27017:27017 -p 1234:1234 sell999
	sudo docker run -d --name filData --network blockchain-browser_frontend -p 9000:9000 sell999
docker-mongo:
	# sudo docker run -d --network host --name mongodb  -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=admin666 -v /opt/gukesoft/middleware/mongo/db:/data/db -v /etc/localtime:/etc/localtime mongo:4.4
	sudo docker run -d -p 27017:27017 --name mongodb  -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=admin666 -v /data/mongo/db:/data/db -v /etc/localtime:/etc/localtime mongo:4.4.6
docker-mysql:
    docker run -p 3306:3306 --name mysql5.7 -v /data/mysql/master/log:/var/log/mysql -v /data/mysql/master/data:/var/lib/mysql  -v /data/mysql/master/conf:/etc/mysql -e MYSQL_ROOT_PASSWORD=root -d bingozhou/mysql5.7
mongodb-exec:
    sudo docker exec -it mongodb mongo -u admin -p admin666 --host localhost:27017 --authenticationDatabase admin
mongodb-backup:
    sudo docker exec -it mongodb mongodump --out /data/db/backup20210706 -u admin -p admin666 --host localhost:27017 --authenticationDatabase admin --db browser --gzip