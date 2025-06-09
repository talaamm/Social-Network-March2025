package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"social-network/internal/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	if db == nil {
		log.Fatal("❌ UserRepository received nil database connection")
	}
	return &UserRepository{DB: db}
}

func (repo *UserRepository) CreateUser(user *models.User) error {
	if repo.DB == nil {
		log.Println("❌ Database connection is nil in CreateUser")
		return sql.ErrConnDone
	}

	_, err := repo.DB.Exec(`
		INSERT INTO users (nickname, email, password, first_name, last_name, age, gender , date_of_birth) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.Nickname, user.Email, user.Password, user.FirstName, user.LastName, user.Age, user.Gender, user.Birthdate)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("email or nickname already in use")
		}
		log.Printf("❌ Failed to insert user into database: %v", err)
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (repo *UserRepository) GetUserByEmailOrNickname(identifier string) (*models.User, string, error) {
	var user models.User
	var storedPassword string

	err := repo.DB.QueryRow(`
		SELECT * FROM users WHERE email = ? OR nickname = ?`, identifier, identifier).
		Scan(&user.ID, &user.Nickname, &user.Email, &storedPassword,
			&user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Birthdate, &user.IsPrivate)
		// fmt.Println()
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("user " + identifier + " doesnt exist")
			return nil, "", nil
		}
		log.Println("❌ Error querying user:", err)
		return nil, "", err
	}
	fmt.Println("user from helper func", user)
	return &user, storedPassword, nil
}

func (repo *UserRepository) GetUserDataById(userID int) (*models.User, error) {
	var user models.User
	err := repo.DB.QueryRow(`
		SELECT id, nickname, email, age, gender, first_name, last_name, date_of_birth , is_private
		FROM users WHERE id = ? `, userID).
		Scan(&user.ID, &user.Nickname, &user.Email,
			&user.Age, &user.Gender, &user.FirstName, &user.LastName,
			&user.Birthdate, &user.IsPrivate)
		// fmt.Println()
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("user " + strconv.Itoa(userID) + " doesnt exist")
			return nil, nil
		}
		log.Println("❌ Error querying user:", err)
		return nil, err
	}
	// fmt.Println("user from helper func", user)
	return &user, nil
}

func (repo *UserRepository) GetUsersNotFollowed(userID int) ([]models.User, error) {
	rows, err := repo.DB.Query(`
        SELECT id, nickname, first_name, last_name, email, age, gender, date_of_birth, is_private 
        FROM users 
        WHERE id NOT IN (
            SELECT following_id FROM followers WHERE follower_id = ?
        ) AND id != ? 
        ORDER BY nickname ASC`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Nickname, &user.FirstName, &user.LastName, &user.Email, &user.Age, &user.Gender, &user.Birthdate, &user.IsPrivate)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
