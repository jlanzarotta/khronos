@echo off
del khronos.exe
::set CGO_ENABLED=1
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
set BUILD_VERSION=1.0.4
set BUILD_DATETIME="%date:~10,4%-%date:~4,2%-%date:~7,2%T%time: =0%"
go build -ldflags="-s -w -X 'khronos/cmd.BuildDateTime=%BUILD_DATETIME%' -X 'khronos/cmd.BuildVersion=%BUILD_VERSION%'" -work -a -v -o khronos.exe main.go
