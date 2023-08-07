package main

// arquitetura:
// - a gente tem usuários pra autenticação. edit: MENTIRA. a gente nem vai ter isso.
// - a gente tem um banco de dados de documentos. CHECAGEM DE FATOS: ✅ tem sim

import (
	"os"

	"hilos/api"
	"hilos/forum"
)

func main() {
	os.MkdirAll("data", os.ModeDir)

	forum.Start()
	forum.RebuildIndex()
	api.Start()
}
