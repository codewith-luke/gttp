package main

type appArgs struct {
	directory string
}

func parseArguments(args []string) appArgs {
	parsedArgs := appArgs{}

	for i, arg := range args {
		if arg == "--directory" {
			value := args[i+1]
			if (i+1) >= len(args) || len(value) == 0 {
				panic("Missing directory argument")
			}

			parsedArgs.directory = value
		}
	}

	return parsedArgs
}
