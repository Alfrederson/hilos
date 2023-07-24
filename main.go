package main

// arquitetura:
// - a gente tem usuários pra autenticação.
// - a gente tem um banco de dados de documentos.

import (
	"fmt"
	"os"

	"plantinha.org/m/v2/api"
	"plantinha.org/m/v2/forum"
)

type Coisa struct {
	X    float64
	Y    float64
	Nome string
}

func (c *Coisa) Describe() string {
	return fmt.Sprintf("%f %f %s", c.X, c.Y, c.Nome)
}

func main() {
	os.MkdirAll("data", os.ModeDir)

	forum.Start()

	api.Start()
}
