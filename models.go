package main

import (
	"context"
	"database/sql"
	"errors"
)

type User struct {
	Id       int64   `json:"id"`
	Name     *string `json:"name,omitempty"`
	Birthday *string `json:"birthday,omitempty"`
	Age      *int    `json:"age,omitempty"'`
	IsMale   *bool   `json:"is_male,omitempty"`
}

var NotFoundError = errors.New("not found")
var NotInsertedError = errors.New("not inserted")
var NotUpdatedError = errors.New("not updated")

func UsersGetAll(ctx context.Context, db *sql.DB) ([]User, error) {
	users := make([]User, 0)
	const query = "SELECT id, name, birthday, age, is_male FROM tb_users"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users, nil
		}
		return nil, err
	}
	for rows.Next() {
		user := User{}
		if err := rows.Scan(&user.Id, &user.Name, &user.Birthday, &user.Age, &user.IsMale); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func UsersGetWithMinAge(ctx context.Context, db *sql.DB, minAge int) ([]User, error) {
	users := make([]User, 0)
	const query = "SELECT id, name, birthday, age, is_male FROM tb_users WHERE age >= $1"
	rows, err := db.QueryContext(ctx, query, minAge)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users, nil
		}
		return nil, err
	}
	for rows.Next() {
		user := User{}
		if err := rows.Scan(&user.Id, &user.Name, &user.Birthday, &user.Age, &user.IsMale); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func UsersGetById(ctx context.Context, db *sql.DB, id int64) (User, error) {
	user := User{}
	const query = "SELECT id, name, birthday, age, is_male FROM tb_users WHERE id = $1"
	row := db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&user.Id, &user.Name, &user.Birthday, &user.Age, &user.IsMale); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, NotFoundError
		}
		return User{}, err
	}

	if err := row.Err(); err != nil {
		return User{}, err
	}
	return user, nil
}

func UsersUpdateById(ctx context.Context, db *sql.DB, user User) error {
	const query = "UPDATE tb_users SET name=$2, birthday=$3, age=$4, is_male=$5 WHERE id=$1"
	res, err := db.ExecContext(ctx, query, user.Id, user.Name, user.Birthday, user.Age, user.IsMale)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected < 1 {
		return NotUpdatedError
	}
	return nil
}

func UsersDeleteById(ctx context.Context, db *sql.DB, id int64) error {
	const query = "DELETE FROM tb_users WHERE id = $1"
	res, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected < 1 {
		return NotFoundError
	}
	return nil
}

func UsersInsert(ctx context.Context, db *sql.DB, user UserData) (int64, error) {
	const query = "INSERT INTO tb_users(name, birthday, age, is_male) VALUES($1, $2, $3, $4) RETURNING id"
	row := db.QueryRowContext(context.Background(), query, user.Name, user.Birthday, user.Age, user.IsMale)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	if err := row.Err(); err != nil {
		return 0, err
	}
	return id, nil
}
