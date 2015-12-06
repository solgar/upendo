#!/bin/sh

versionStr="$(git show-ref refs/heads/master --hash | tr -d "\n")"
basedOnVersion="$(git show-ref refs/remotes/origin/master --hash | tr -d "\n")"
untrackedCount="$(git ls-files --others --exclude-standard | wc -l)"
modifiedCount="$(git ls-files -m | wc -l)"

printf "package main\n\n" > version.go
printf "var (\n" >> version.go
printf "\tVersion                 = \"$versionStr\"\n" >> version.go
printf "\tBasedOnVersion          = \"$basedOnVersion\"\n" >> version.go
printf "\tUntrackedFilesCount int = $untrackedCount\n" >> version.go
printf "\tModifiedFilesCount  int = $modifiedCount\n" >> version.go
printf ")\n" >> version.go
