# A chord implementation in golang

To get started
-----
1. set GOPATH: export GOATH=~/path/to/gopath
2. mkdir -p $GOPATH/src/github.com/hoffa2/chord
3. cd $GOPATH/src/github.com/hoffa2/src
4. git clone github.com/hoffa2/chord
5. To persist the GOPATH setting add the same line to either your ~/.bashrc or ~/.zshrc

Convenience
-----
* To download all dependencies - just type **go get ./..** in the folder containing your source code
How to run it
-----
* To run either of the components: Type **go run main.go name-of-component arguments** (By just typing go run main.go - the cli will give information about parameters)
* Example: go run main.go runall --nameserver=compute-1-3 --graph=1
* once in the cli type: **add** to add a random node to the ring. to test: type **test total-number-of-requests**


Command in organizer cli
-----
| Command       | function          | args  |
| ------------- |:-------------:| -----:|
| ls            | lists current nodes | 0 |
| add           | adds a random node      |    |
| leave         | removes a node from the network | the name of the node |
| test          | runs the client|  number of requests |
| CTRL-C or shutdown | closes all connections and the cli | 0 |
