package database

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrDatabaseConnection = errors.New("error connecting to the database")
	ErrDatabaseUpdateUser = errors.New("error updating user in database")
	ErrDatabaseGetUser    = errors.New("error get user(s) in database")
	ErrDatabaseDeleteUser = errors.New("error deleting user in database")
)

type User struct {
	MitIdUUID string //MitIdUUID
	GaiaId    string `gorm:"primaryKey"` //Business ID
	Name      string //Full name
	Address   string
	DarId     string
	Updated   int64 `gorm:"autoUpdateTime"`
	Created   int64 `gorm:"autoCreateTime"`
}

type UserDatabase struct {
	db *gorm.DB
}

func New(host string) *UserDatabase {
	db, err := gorm.Open(sqlite.Open(host), &gorm.Config{Logger: logger.Default.LogMode((logger.Warn))})
	if err != nil {
		panic(errors.Join(ErrDatabaseConnection, err))
	}

	db.AutoMigrate(&User{})
	return &UserDatabase{
		db: db,
	}
}

func (db *UserDatabase) GetUserById(userId string) (user User, err error) {
	result := db.db.Find(&user, "gaia_id = ?", userId)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		err = errors.Join(ErrDatabaseGetUser, result.Error)
	}
	return user, err
}

func (db *UserDatabase) GetUsers() (users []User, err error) {
	result := db.db.Find(&users)
	if result.Error != nil {
		return users, errors.Join(ErrDatabaseGetUser, result.Error)
	}
	return users, err
}

func (db *UserDatabase) UpsertUser(user User) (err error) {
	_, err = db.BulkUpsertUsers([]User{user})
	return err
}

func (db *UserDatabase) BulkUpsertUsers(users []User) (rows int64, err error) {
	result := db.db.Save(&users)
	if result.Error != nil {
		return rows, errors.Join(ErrDatabaseUpdateUser, result.Error)
	}
	rows = int64(result.RowsAffected)
	return rows, err
}

func (db *UserDatabase) DeleteUser(userId string) (err error) {
	result := db.db.Delete(&User{}, userId)
	if result.Error != nil {
		err = errors.Join(ErrDatabaseDeleteUser, result.Error)
	}
	return
}
