/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/skye-lopez/go-get-cli/cmd"
	"github.com/skye-lopez/go-get-cli/store"
)

func _main() {
	cmd.Execute()
}

// Testing store
func main() {
	store.Init()
}
