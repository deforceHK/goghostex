package main

import "os"

func main() {

	cmd := &Command{}
	cmd.New(os.Args)

	if err := cmd.Parse(); err != nil {
		panic(err)
	}
}
