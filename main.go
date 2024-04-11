package main

import (
	"fmt"
	"hilos/api"
	"hilos/forum"
	"hilos/identity"
	"os"
)

// arquitetura:
// - a gente tem usuários pra autenticação. edit: MENTIRA. a gente nem vai ter isso.
// - a gente tem um banco de dados de documentos. CHECAGEM DE FATOS: ✅ tem sim

func main() {
	i := identity.New()
	i.Name = "ADM"
	i.Powers = 95
	i.Sign()
	encoded, err := i.EncodeBase64()
	fmt.Println(err)
	fmt.Println(encoded)

	os.MkdirAll("data", os.ModeDir)
	forum.Start()
	api.Start()
}
