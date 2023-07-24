package identity

import (
	"testing"
)

func TestIdentity_Check(t *testing.T) {
	i := New()
	i.Sign()

	if !i.Check() {
		t.Errorf("checagem de assinatura válida falhou: %s ", i.Signature)
	}

	i.Signature = "assinatura inválida"

	if i.Check() {
		t.Errorf("checagem de assinatura inválida deveria ter falhado: %s ", i.Signature)
	}

}

func TestDecodeBase64(t *testing.T) {
	i := New()

	encoded, _ := i.EncodeBase64()

	d, err := DecodeBase64(encoded)
	if err != nil {
		t.Errorf("falhou em decodificar identidade válida")
	}

	if d == nil {
		t.Errorf("a identidade não deveria ser nil")
	}

	if d.Name != i.Name {
		t.Errorf("a identidade decodificada sofreu mutação %s x %s", d.Name, i.Name)
	}
	if d.Id != i.Id {
		t.Errorf("a identidade decodificada sofreu mutação %s x %s", d.Id, i.Id)
	}
	if d.Powers != i.Powers {
		t.Errorf("a identidade decodificada sofreu mutação %d x %d", d.Powers, i.Powers)
	}
	if d.Signature != d.Signature {
		t.Errorf("a identidade decodificada sofreu mutação %s x %s", d.Signature, i.Signature)
	}
	_, err = DecodeBase64("obviamente inválido")
	if err == nil {
		t.Errorf("decodificar identidade inválida deveria falhar")
	}

}
