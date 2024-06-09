EXEC=GithubIssues
include .env

all: $(EXEC)

$(EXEC): client/dist/index.html keys
	go build -o $(EXEC) .

client/dist/index.html:
	git submodule update --remote
	npm i --prefix client/
	echo VITE_CLIENT_ID=${CLIENT_ID} > client/.env
	npm  run build --prefix client/ 
	rm -f client/.env

keys:
	mkdir keys
	ssh-keygen -f keys/key.pem -m pkcs8 -N ""
	ssh-keygen -f keys/key.pem.pub -e -m pkcs8 > keys/key.pub
	rm -f keys/key.pem.pub

clean:
	rm -f $(EXEC) 
	rm -rf keys
	rm -rf client/dist
