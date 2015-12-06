#!/bin/sh

printf "package main

var (
	Version             string = \"unknown\"
	BasedOnVersion      string = \"unknown\"
	UntrackedFilesCount int    = -1
	ModifiedFilesCount  int    = -1
)
" > version.go
