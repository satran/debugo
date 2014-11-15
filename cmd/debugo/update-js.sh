#!/bin/bash

# The point of this script is so that I can embedd all js files within the go executable.

jquery="http://code.jquery.com/jquery-2.1.1.min.js"
underscore="http://underscorejs.org/underscore-min.js"
backbone="http://backbonejs.org/backbone-min.js"

echo -e "package main\nvar jquery=\`" > jquery.go
curl $jquery >> jquery.go
echo "\`" >> jquery.go

echo -e "package main\nvar underscore=\`" > underscore.go
curl $underscore >> underscore.go
echo "\`" >> underscore.go

echo -e "package main\nvar backbone=\`" > backbone.go
curl $backbone >> backbone.go
echo "\`" >> backbone.go



