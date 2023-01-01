package cli

import "flag"

type Arguments struct {
	Quiet   bool
	Variant string

	Services []string
}

func ParseArguments() Arguments {
	var flags Arguments

	flag.BoolVar(&flags.Quiet, "q", false, "make logging quiet")
	flag.StringVar(&flags.Variant, "variant", "", "specify the system variant")
	flag.Parse()
	flags.Services = flag.Args()
	return flags
}
