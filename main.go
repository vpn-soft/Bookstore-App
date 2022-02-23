package main

import api_server "github.com/Bookstore-App/api-server"

func main() {
	port_number := "8888"
	api_server.StartAPIServer(port_number)
}
