# Chat #

## Warning: you should use 'go build runServer.go server.go', './runServer' in /server dirrectory for starting server. 'Go run runServer.go' doesn`t work because we have 2 files in package main ##

commands:
`publish <room> : <message>`
`subscribe <room> : <nickname>`

You get all previous room history(or 128 last messages, if there were more then 128 of them) when subscribing. After that every new message from anybody else in this room will be shown as soon as it will be published.