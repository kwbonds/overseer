package main

type stringsFlag []string

func (i *stringsFlag) String() string {
	return "strings array"
}

func (i *stringsFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}
