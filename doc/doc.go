package doc

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
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

type Doc struct {
	CreatedAt time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"updated_at"`

	Path string `json:"path" gorm:"primaryKey"`
	Data datatypes.JSON
}

type Indexable = interface {
	IndexTable() interface{}
	IndexedFields() interface{}
	ObjectIndex() []string
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
	Name      string
	conn      *gorm.DB
	mutex     sync.Mutex
	txMutex   sync.Mutex
	Type      int
	indexable Indexable
}

func New(parts ...interface{}) string {
	if len(parts) == 0 {
		return GenerateId(16)
	}
	strParts := make([]string, len(parts))
	for i, part := range parts {
		strParts[i] = fmt.Sprint(part)
	}
	joined := strings.Join(strParts, "/")
	generatedID := GenerateId(10)
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
		UpdatedAt: time.Now(),
		Path:      path,
		Data:      bytes,
	}
	log.Println("rewriting " + path)
	db.conn.Save(&doc)
	return nil
}

func (db *DocDB) Clear() {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.conn.Exec("DELETE from docs")
}

func (db *DocDB) Add(path string, object interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	bytes, err := json.Marshal(object)
	if err != nil {
		log.Println(err)
		return errors.New("invalid object")
	}
	// grava o documento...
	db.conn.Debug().Exec("INSERT INTO docs(path,data,created_at,updated_at) VALUES(?,?,?,?)", path, string(bytes), time.Now(), time.Now())
	return nil
}

func (db *DocDB) Exists(path string) bool {
	type Path struct {
		Path string `gorm:"primaryKey"`
	}
	tx := db.conn.Table("docs").First(&Path{}, &Path{
		Path: path,
	})
	return tx.Error == nil
}

func (db *DocDB) Get(path string, object interface{}) error {
	result := Doc{}
	db.conn.First(&result, &Doc{
		Path: path,
	})
	if result.Path == "" {
		return errors.New("document no found: " + path)
	}
	if object != nil {
		json.Unmarshal(result.Data, object)
	}
	return nil
}

func (db *DocDB) Delete(path string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.conn.Delete(path)
}

// Pega os últimos perPage documentos da página page por ordem de atualização
func (db *DocDB) GetLastUpdated(page int, perPage int) ([]string, error) {
	var stuff = make([]string, 0, perPage)
	db.conn.
		Table("docs").
		Select("data").
		Offset(page * perPage).
		Limit(perPage).
		Order("docs.updated_at DESC").
		Find(&stuff)
	return stuff, nil
}

// Retorna o JSON de dentro dos objectos, ao invés de retornar os próprios objectos.
func (db *DocDB) FindLastUpdated(field string, op string, value any, page int, perPage int) ([]string, error) {
	// Where("data->>'$."+field+"' = ?", value).
	var stuff = make([]string, 0, perPage)
	db.conn.
		Table("docs").
		Select("data").
		Where("data->>'$."+field+"' "+op+" ?", value).
		Offset(perPage * page).
		Limit(perPage).
		Order("docs.updated_at DESC").
		Find(&stuff)
	return stuff, nil
}

func (db *DocDB) FindLast(field string, op string, value any, page int, perPage int) ([]string, error) {
	var stuff = make([]string, 0, perPage)
	db.conn.Debug().
		Table("docs").
		Select("data").
		Where("data->>'$."+field+"' "+op+" ?", value).
		Offset(page * perPage).
		Limit(perPage).
		Order("docs.created_at DESC").
		Find(&stuff)
	return stuff, nil
}

func (db *DocDB) Find(field string, op string, value string, page int, perPage int) ([]string, error) {
	//SELECT * FROM employees WHERE address->>'$.postalCode' = '60611';
	type Entry struct {
		Path string
	}
	docs := make([]Entry, 0, perPage)
	// INJECTION!!!
	db.conn.Debug().
		Table("docs").
		Select("path").
		Where("data->>'$."+field+"' = ?", value).
		Offset(page * perPage).
		Limit(perPage).
		Find(&docs)
	result := make([]string, 0, len(docs))
	for _, entry := range docs {
		result = append(result, entry.Path)
	}
	return result, nil
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

func Create(file string, indexable Indexable) *DocDB {
	conn := sqlite.Open(DOCDB_PATH + file)

	result := DocDB{
		Name:      file,
		Type:      TYPE_COLLECTION,
		indexable: indexable,
	}
	var err error

	result.conn, err = gorm.Open(conn, &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	// result.conn.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		panic("não consegui criar instancia do docdb")
	}
	result.conn.AutoMigrate(&Doc{})
	return &result
}

func dropAllIndexes(db *gorm.DB) error {
	var tables []string
	if err := db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Pluck("name", &tables).Error; err != nil {
		return err
	}
	for _, table := range tables {
		var indexes []struct {
			Name string
		}
		if err := db.Raw(fmt.Sprintf("PRAGMA index_list(%s)", table)).Scan(&indexes).Error; err != nil {
			return err
		}
		for _, index := range indexes {
			if err := db.Exec(fmt.Sprintf("DROP INDEX %s", index.Name)).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DocDB) RebuildIndex() {
	log.Println("rebuilding index for ", d.Name)
	// drop indices
	var tableNames []string
	log.Println("dropando indices sistema antigo...")
	result := d.conn.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 't_%'").Scan(&tableNames)
	if result.Error != nil {
		panic(result.Error)
	}
	for _, tableName := range tableNames {
		result := d.conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
		if result.Error != nil {
			panic(result.Error)
		}
		fmt.Printf("dropped %s \n", tableName)
	}
	indexModel := d.indexable.IndexTable()
	if err := d.conn.Migrator().DropTable(indexModel); err != nil {
		log.Println("erro dropando indice: ", err)
	}
	// dropa todos os índices
	dropAllIndexes(d.conn)
	// recria eles
	for _, index := range d.indexable.ObjectIndex() {
		d.conn.Debug().Exec("CREATE INDEX idx_" + index + " ON docs((data->>'$." + index + "'))")
	}

	d.conn.Debug().Exec("UPDATE docs SET data = JSON_SET(COALESCE(data, '{}'), '$.parent_id', '') WHERE data->>'$.parent_id' IS NULL")
	// o tempo todo era só ter feito isso
	//CREATE INDEX idx_postal_code ON employees((address->>'$.postalCode'));
}

func init() {
	DOCDB_PATH = os.Getenv("DOCDB_PATH")
	log.Println("initializing docdb, path = ", DOCDB_PATH)
}
