package main

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string
}

func createUser(db *gorm.DB, user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	result := db.Create(user)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func login(db *gorm.DB, user *User) (string, error) {
	targetUser := new(User)
	result := db.Where("email = ?", user.Email).First(targetUser)

	if result.Error != nil {
		return "", result.Error
	}

	err := bcrypt.CompareHashAndPassword([]byte(targetUser.Password), []byte(user.Password))

	if err != nil {
		return "", err
	}

	jwtSecret := "Test"
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = targetUser.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", err
	}

	return t, nil
}
