package main

type Compiler struct {
	Sc *Scanner
}

func (c *Compiler) compile(source string) {
	c.Sc.initScanner(source)
}
