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
	"gorm.io/gorm/logger"
)

var DOCDB_PATH string

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const ALL = 0

func CreateNewObject(obj interface{}) interface{} {
	objType := reflect.TypeOf(obj)
	newObj := reflect.New(objType.Elem()).Interface()
	return newObj
}

func GenerateId(length int) string {
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

// Gera um caminho de documento a partir de um prefixo.
// retorna algo tipo coisas/quadradas/ASDOIJZXCXC
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

// Salva um documento no caminho especificado. Isso equivale a um upsert.
func (db *DocDB) Save(path string, object interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	bytes, err := json.Marshal(object)
	if err != nil {
		log.Println(err)
		return errors.New("invalid object")
	}

	// código comentado vai ficar aqui por razões históricas, mas ele está aqui porque
	// o gorm estava sumindo com a coluna created_at
	// doc := Doc{
	// 	UpdatedAt: time.Now(),
	// 	Path:      path,
	// 	Data:      bytes,
	// }

	now := time.Now()
	db.conn.Exec(
		"INSERT INTO docs(path,created_at,updated_at,data) VALUES (?,?,?,?) ON CONFLICT (path) DO UPDATE SET updated_at=excluded.updated_at, data=excluded.data",
		path,
		now,
		now,
		string(bytes),
	)
	return nil
}

func (db *DocDB) Clear() {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.conn.Exec("DELETE from docs")
}

// Adiciona um documento no caminho especificado.
func (db *DocDB) Add(path string, object interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	bytes, err := json.Marshal(object)
	if err != nil {
		log.Println(err)
		return errors.New("invalid object")
	}
	// grava o documento...
	db.conn.Exec("INSERT INTO docs(path,data,created_at,updated_at) VALUES(?,?,?,?)", path, string(bytes), time.Now(), time.Now())
	return nil
}

// Diz se um documento existe com o caminho especificado.
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
		return errors.New("document not found: " + path)
	}
	if object != nil {
		json.Unmarshal(result.Data, object)
	}
	return nil
}

func (db *DocDB) Delete(path string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.conn.Delete(&Doc{
		Path: path,
	})
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

// Pega os últimos perPage documentos da página page por ordem de criação
func (db *DocDB) GetLastCreated(page int, perPage int) ([]string, error) {
	var stuff = make([]string, 0, perPage)
	db.conn.
		Table("docs").
		Select("data").
		Offset(page * perPage).
		Limit(perPage).
		Order("docs.created_at DESC").
		Find(&stuff)
	return stuff, nil
}

type Condition struct {
	Field string
	Op    string
	Value any
}

// conta quantos registros tem na coleção.
func (db *DocDB) Count() int64 {
	var r int64 = 0
	db.conn.Table("docs").Count(&r)
	return r
}

// conta quantos registros tem de acordo com as condições.
func (db *DocDB) CountWhere(conditions ...Condition) int64 {
	var r int64 = 0
	tx := db.conn.Table("docs")
	for _, cond := range conditions {
		tx = tx.Where("data->>'$."+cond.Field+"' "+cond.Op+" ?", cond.Value)
	}
	tx.Count(&r)
	return r
}

// lista os perPage ids a partir de page*perPage
func (db *DocDB) List(page int, perPage int) ([]string, error) {
	stuff := make([]string, 0)
	tx := db.conn.Table("docs").Select("path")
	if perPage != ALL {
		tx = tx.Offset(perPage * page).Limit(perPage)
	}
	tx.Order("docs.created_at ASC").Find(&stuff)
	return stuff, nil
}

// Retorna os caminhos (ids) dos documentos onde condição por ordem de criação.
// page é o offset.
// per page são quantos documentos retornar por página.
// se per page for zero, manda todos os que eu achar.
func (db *DocDB) ListWhere(page int, perPage int, conditions ...Condition) ([]string, error) {
	stuff := make([]string, 0)
	tx := db.conn.Table("docs").Select("path")
	for _, cond := range conditions {
		tx = tx.Where("data->>'$."+cond.Field+"' "+cond.Op+" ?", cond.Value)
	}
	if perPage != ALL {
		tx = tx.Offset(perPage * page).Limit(perPage)
	}
	tx.Order("docs.created_at ASC").Find(&stuff)
	return stuff, nil
}

// pega os primeiros perPage itens por ordem de criação.
func (db *DocDB) FindFirstCreated(page int, perPage int, conditions ...Condition) ([]string, error) {
	var stuff = make([]string, 0)
	tx := db.conn.Table("docs").Select("data")
	for _, cond := range conditions {
		tx = tx.Where("data->>'$."+cond.Field+"' "+cond.Op+" ?", cond.Value)
	}
	if perPage != ALL {
		tx = tx.Offset(perPage * page).Limit(perPage)
	}
	tx.Order("docs.created_at ASC").Find(&stuff)
	return stuff, nil
}

// pega os últimos perPage itens por ordem de criação.
func (db *DocDB) FindLastCreated(page int, perPage int) ([]string, error) {
	var stuff = make([]string, 0)
	tx := db.conn.Table("docs").Select("data")
	if perPage != ALL {
		tx = tx.Offset(perPage * page).Limit(perPage)
	}
	tx.Order("docs.created_at DESC").Find(&stuff)
	return stuff, nil
}

func (db *DocDB) FindFirstCreatedWhere(page int, perPage int, conditions ...Condition) ([]string, error) {
	var stuff = make([]string, 0)
	tx := db.conn.Table("docs").Select("data")
	for _, cond := range conditions {
		tx = tx.Where("data->>'$."+cond.Field+"' "+cond.Op+" ?", cond.Value)
	}
	if perPage != ALL {
		tx = tx.Offset(perPage * page).Limit(perPage)
	}
	tx.Order("docs.created_at ASC").Find(&stuff)
	return stuff, nil
}

// manda os últimos perPage itens por ordem de criação.
func (db *DocDB) FindLastCreatedWhere(page int, perPage int, conditions ...Condition) ([]string, error) {
	var stuff = make([]string, 0)
	tx := db.conn.Table("docs").Select("data")
	for _, cond := range conditions {
		tx = tx.Where("data->>'$."+cond.Field+"' "+cond.Op+" ?", cond.Value)
	}
	if perPage != ALL {
		tx = tx.Offset(perPage * page).Limit(perPage)
	}
	tx.Order("docs.created_at DESC").Find(&stuff)
	return stuff, nil
}

// Pega o JSON dos últimos documentos atualizados onde algumas condições são verdadeiras.
// page é o offset.
// per page são quantos documentos retornar por página.
// se per page for zero, manda todos os que eu achar.
func (db *DocDB) FindLastUpdatedWhere(page int, perPage int, conditions ...Condition) ([]string, error) {
	var stuff = make([]string, 0)
	tx := db.conn.Table("docs").Select("data")
	for _, cond := range conditions {
		tx = tx.Where("data->>'$."+cond.Field+"' "+cond.Op+" ?", cond.Value)
	}
	if perPage != ALL {
		tx = tx.Offset(perPage * page).Limit(perPage)
	}
	tx.Order("docs.updated_at DESC").Find(&stuff)
	return stuff, nil
}

// Retorna o JSON de dentro dos registros que foram atualizados mais recentemente.
func (db *DocDB) FindLastUpdated(page int, perPage int) ([]string, error) {
	// Where("data->>'$."+field+"' = ?", value).
	var stuff = make([]string, 0, perPage)
	tx := db.conn.Table("docs").Select("data")
	if perPage > 0 {
		tx = tx.Offset(perPage * page).Limit(perPage)
	}
	tx.Order("docs.updated_at DESC").Find(&stuff)
	return stuff, nil
}

// Pega o ID de todos os documentos que atendam um critério.
// ex: pra pegar o que tem { "cor" : "laranja" }, eu uso bolinhas.Find("cor","=","laranja",0,0)
func (db *DocDB) Find(field string, op string, value string, page int, perPage int) ([]string, error) {
	//SELECT * FROM employees WHERE address->>'$.postalCode' = '60611';
	type Entry struct {
		Path string
	}
	docs := make([]Entry, 0, perPage)
	// INJECTION!!! <- medo infundado?
	tx := db.conn.Table("docs").Select("path").Where("data->>'$."+field+"' = ?", value)
	if perPage > 0 {
		tx = tx.Offset(page * perPage).Limit(perPage)
	}
	tx.Find(&docs)
	result := make([]string, 0, len(docs))
	for _, entry := range docs {
		result = append(result, entry.Path)
	}
	return result, nil
}

// pra fazer algumas coisas malucas...
func (db *DocDB) Conn() *gorm.DB {
	return db.conn
}

func (db *DocDB) Begin() {
	db.txMutex.Lock()
	db.conn.Exec("BEGIN TRANSACTION")
}

func (db *DocDB) Commit() {
	db.conn.Exec("COMMIT")
	db.txMutex.Unlock()
}

func (db *DocDB) Rollback() {
	db.conn.Exec("ROLLBACK")
	db.txMutex.Unlock()
}

// transforma uma array de JSONS em uma array de T.
func RecordsToStructs[T any](records []string) []*T {
	result := make([]*T, 0, len(records))
	for _, v := range records {
		var entry T
		err := json.Unmarshal([]byte(v), &entry)
		if err != nil {
			log.Println("RecordsToStructs:", err)
		}
		result = append(result, &entry)
	}
	return result
}

type tranny []*DocDB

func (t tranny) Begin() {
	for _, val := range t {
		val.Begin()
	}
}
func (t tranny) Commit() {
	for _, val := range t {
		val.Commit()
	}
}
func (t tranny) Rollback() {
	for _, val := range t {
		val.Rollback()
	}
}

// returns a tranny
func Tranny(docs ...*DocDB) tranny {
	return tranny(docs)
}

// Cria uma nova "instância" de uma coleção de um tipo específico.
func Create(file string, indexable Indexable, logLevel logger.LogLevel) *DocDB {
	conn := sqlite.Open(DOCDB_PATH + file)

	result := DocDB{
		Name:      file,
		Type:      TYPE_COLLECTION,
		indexable: indexable,
	}
	var err error

	result.conn, err = gorm.Open(conn, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
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

// Reconstroi o índice da coleção.
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
