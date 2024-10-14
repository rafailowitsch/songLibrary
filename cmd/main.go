package main

import (
	"songLibrary/internal/app"
)

// @title Song Library API
// @version 1.0
// @description This is a sample server for managing songs.
// @host localhost:8089
// @BasePath /
// @schemes http
func main() {
	app.Run()
}
