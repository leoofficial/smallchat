# Smallchat

Inspired by Salvatore's [smallchat](https://github.com/antirez/smallchat), I implemented a Go version,
supporting TCP and UDP connection. 

Run the server and chat using [ncat](https://nmap.org/ncat/).

TCP Server
```
ncat localhost 3000
```
UDP Server
```
ncat -u localhost 3000
```