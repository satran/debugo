# Debugo

This is a very naive debugger written in Go for Go. It works only on Linux and on an x86(tested only on a x86-64) architecture. I started this to learn more on the internals of both Go and a bit about Linux. 

Right now it just traces through the main function and displays the lines that got executed. The project is split as a command and a library. The library can be reused for building similar debuggers.
