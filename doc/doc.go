package doc

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"

	"encoding/json"
	"log"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DOCDB_PATH string

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func CreateNewObject(obj interface{}) interface{} {
	objType := reflect.TypeOf(obj)
	newObj := reflect.New(objType.Elem()).Interface()
	return newObj
}

func GenerateId(length int) string {
	rand.Seed(time.Now().UnixNano())
	randomBytes := make([]byte, length)
	for i := 0; i < length; i++ {
		randomBytes[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomBytes)
}

var db *gorm.DB

type Doc struct {
	Path string `gorm:"primaryKey"`
	Data datatypes.JSON
}

type Indexable = interface {
	ReadField(string) (string, error)
}

type DocumentData map[string]interface{}

func (d DocumentData) ToJsonBytes() ([]byte, error) {
	bytes, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (d *DocumentData) FromJsonBytes(data []byte) {
	if err := json.Unmarshal(data, d); err != nil {
		log.Println("Erro decodificando: ", err)
	}
}

const (
	TYPE_COLLECTION = 0
	TYPE_INDEX      = 1
)

type DocDB struct {
	conn    *gorm.DB
	mutex   sync.Mutex
	txMutex sync.Mutex
	Type    int
}

func New(parts ...interface{}) string {
	if len(parts) == 0 {
		return GenerateId(16)
	}
	// Convert each argument to a string
	strParts := make([]string, len(parts))
	for i, part := range parts {
		strParts[i] = fmt.Sprint(part)
	}

	// Join the parts with slashes
	joined := strings.Join(strParts, "/")

	// Generate a unique ID (example: using time.Now().UnixNano())
	generatedID := GenerateId(10) // Replace with your actual generated ID logic

	// Append the generated ID to the joined string
	result := fmt.Sprintf("%s/%s", joined, generatedID)

	return result
}

func (db *DocDB) Save(path string, object interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	bytes, err := json.Marshal(object)
	if err != nil {
		log.Println(err)
		return errors.New("invalid object")
	}
	doc := Doc{
		Path: path,
		Data: bytes,
	}

	db.conn.Save(&doc)

	return nil
}

func (db *DocDB) Add(path string, object interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	bytes, err := json.Marshal(object)
	if err != nil {
		log.Println(err)
		return errors.New("invalid object")
	}
	db.conn.Exec("INSERT INTO docs(path,data) VALUES(?,?)", path, string(bytes))
	return nil
}

func (db *DocDB) Get(path string, object interface{}) error {
	result := Doc{}
	db.conn.First(&result, &Doc{
		Path: path,
	})
	if result.Path == "" {
		return errors.New("document no found: " + path)
	}
	json.Unmarshal(result.Data, object)
	return nil
}

func (db *DocDB) Delete(path string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.conn.Delete(path)
}

func (db *DocDB) List(path string, from int, limit int) []string {
	if db.Type == TYPE_INDEX {
		type Doc struct {
			Data string
		}
		docs := make([]Doc, 0, 10)
		db.conn.Where("path = ?", path).Offset(from).Limit(limit).Find(&docs)
		list := make([]string, 0, 10)
		for _, doc := range docs {
			list = append(list, doc.Data[1:len(doc.Data)-1])
		}
		return list
	} else {
		list := []string{"list operation can only be useed on indices"}
		return list
	}
}

func (db *DocDB) Begin() {
	db.txMutex.Lock()
	db.conn.Exec("BEGIN")
}

func (db *DocDB) Commit() {
	db.conn.Exec("COMMIT")
	db.txMutex.Unlock()
}

func (db *DocDB) Rollback() {
	db.conn.Exec("ROLLBACK")
	db.txMutex.Unlock()
}

func Begin(group ...*DocDB) {
	for _, v := range group {
		v.Begin()
	}
}
func Commit(group ...*DocDB) {
	for _, v := range group {
		v.Commit()
	}
}
func Rollback(group ...*DocDB) {
	for _, v := range group {
		v.Rollback()
	}
}

func Create(file string) *DocDB {
	conn := sqlite.Open(DOCDB_PATH + file)

	result := DocDB{
		Type: TYPE_COLLECTION,
	}
	var err error

	result.conn, err = gorm.Open(conn, &gorm.Config{})
	result.conn.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		panic("não consegui criar instancia do docdb")
	}
	result.conn.AutoMigrate(&Doc{})

	return &result
}

// Isso tem que ser um método dentro do DocDB que aceite como parâmetro os campos usados pra indexação.
/*
	ex: db.CreateIndex(Campos{"creator_id","parent_id"})

	- isso vai percorrer Campos{...} criando uma tabela de índice pra cada valor.
	- sempre que um documento for salvo, ele vai chamar um método dentro do struct chamado Get("creator_id") ou Get("parent_id"), que vai retornar o valor I.
	- ele vai então criar uma entrada na tabela correspondente com aquele valor e o ID do documento.

*/

func (d *DocDB) UsingIndexable(i Indexable) {
	log.Println("Using indexable")
}

func CreateIndex(file string) *DocDB {
	conn := sqlite.Open(DOCDB_PATH + file)
	result := DocDB{
		Type: TYPE_INDEX,
	}
	type Doc struct {
		//gorm.Model
		Path string `gorm:"index"`
		Data datatypes.JSON
	}

	var err error
	result.conn, err = gorm.Open(conn, &gorm.Config{})
	if err != nil {
		panic("não consegui criar instância do docdb")
	}
	result.conn.AutoMigrate(&Doc{})
	return &result
}

func init() {
	DOCDB_PATH = "data" // os.Getenv("DOCDB_PATH")
	log.Println("initializing docdb, path = ", DOCDB_PATH)

}
