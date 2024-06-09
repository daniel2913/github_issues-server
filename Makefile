EXEC=GithubIssues 

all: $(EXEC)

$(EXEC): keys client
	go build -o $(EXEC) .

client:
	npm run --prefix client/ build

keys:
	mkdir keys
	ssh-keygen -f keys/key.pem -m pkcs8 -N ""
	ssh-keygen -f keys/key.pem.pub -e -m pkcs8 > keys/key.pub
	rm -f keys/key.pem.pub

clean:
	rm -f$(EXEC) 
	rm -rf keys

