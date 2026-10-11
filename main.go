package main

func main() {
	LoadConfig()
	LoadBlankInfos()
	go RequestLoop()
	API()
}
