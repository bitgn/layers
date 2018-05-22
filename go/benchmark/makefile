PROJ:=github.com/bitn/layers
VAULT:=VAULT
TAG:=`git describe --tags`
LDFLAGS:=-X main.buildVersion=$(TAG)



build:
	go build -ldflags "$(LDFLAGS)" -o benchcli

remote:

	ssh vault PROJ=$PROJ 'bash -s' <<'ENDSSH'
		set -e
		cd $GOPATH/src/$PROJ
		git pull
		make build
		mv benchcli /tmp
	ENDSSH
	scp $VAULT:/tmp/benchcli .

