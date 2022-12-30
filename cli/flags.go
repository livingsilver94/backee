package cli

import "flag"

type Flags struct {
	Variant string
}

func ParseFlags() Flags {
	var flags Flags

	flag.StringVar(&flags.Variant, "variant", "", "specify the system variant")
	flag.Parse()
	return flags
}
