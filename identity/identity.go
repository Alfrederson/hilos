package identity

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"encoding/json"

	"hilos/doc"
)

func sign(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))
	return signature
}

type Identity struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	IP        string `json:"-"`
	Powers    int    `json:"powers"`
	Signature string `json:"sign"`
}

var SECRET string

func (i *Identity) Check() bool {
	text := fmt.Sprintf("%s %s %d salzinho", i.Name, i.Id, i.Powers)
	if sign(text, SECRET) != i.Signature {
		return false
	}
	return true
}

func (i *Identity) Sign() {
	text := fmt.Sprintf("%s %s %d salzinho", i.Name, i.Id, i.Powers)
	i.Signature = sign(text, SECRET)
}

func (i *Identity) EncodeBase64() (string, error) {
	jsonified, err := json.Marshal(i)
	if err != nil {
		return "", errors.New("couldn't encode identity as base64")
	}
	return base64.StdEncoding.EncodeToString(jsonified), nil
}

func DecodeBase64(encoded string) (*Identity, error) {
	if encoded == "" {
		return nil, errors.New("no identity provided")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		log.Println("Erro decodificando identidade: ", err)
		return nil, errors.New("could not decode identity")
	}

	i := Identity{}
	err = json.Unmarshal(decoded, &i)
	if err != nil {
		log.Println("Erro decodificando json da identidade: ", err)
		return nil, errors.New("could not parse identity")
	}
	return &i, nil
}

func makeName() string {
	rand.Seed(time.Now().UnixNano())
	animalIndex := rand.Intn(len(animais))
	adjectiveIndex := rand.Intn(len(adjetivos))
	year := rand.Intn(101) + 1900

	animal := animais[animalIndex]
	adjective := adjetivos[adjectiveIndex]

	return fmt.Sprintf("%s %s %d-%s", animal, adjective, year, doc.GenerateId(4))

}

func New() Identity {
	i := Identity{
		Name:   makeName(),
		Id:     doc.GenerateId(20),
		Powers: 1,
	}
	i.Sign()
	return i
}

func SetSecret(s string) {
	SECRET = s
	log.Println("changing secret")
}

func init() {
	SECRET = os.Getenv("RWT_SECRET")
	if SECRET == "" {
		SECRET = "POR FAVOR FALSIFIQUEM MEUS TOKENS"
		log.Println(SECRET)
	}
}
