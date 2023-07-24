package users

import (
	"log"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	gorm.Model

	Name     string
	Email    string
	Password string
	Claims   datatypes.JSON
}

func Create(u *User) {
	db.Create(u)
}

func Get(id uint) *User {
	result := User{}
	db.First(&result, 1)
	return &result
}

func Delete(u *User) {
	db.Delete(&u)
}

func All() []User {
	var users []User
	db.Find(&users)
	return users
}

func Start(conn gorm.Dialector) {
	var err error

	db, err = gorm.Open(conn, &gorm.Config{})
	if err != nil {
		panic("não consegui conectar no banco de dados de usuários")
	}

	db.AutoMigrate(&User{})

	log.Println("users component activated")
}
