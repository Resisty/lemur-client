#!/usr/bin/env bash
OLDGOPATH=${OLDGOPATH:-$GOPATH}
OLDPS1=${OLDPS1:-$PS1}
export GOPATH=$(pwd)
export PS1="($(basename $PWD)) $PS1"
go get -u github.com/kardianos/govendor
GOVENDOR=$GOPATH/bin/govendor
cd src/lemur/
$GOVENDOR init
$GOVENDOR fetch +missing
cd $GOPATH

function revert() {
	export GOPATH=${OLDGOPATH}
	export PS1="${OLDPS1}"
}

