package cli

import "flag"

type Arguments struct {
	Variant  string
	Services []string
}

func ParseArguments() Arguments {
	var flags Arguments

	flag.StringVar(&flags.Variant, "variant", "", "specify the system variant")
	flag.Parse()
	flags.Services = flag.Args()
	return flags
}
