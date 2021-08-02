package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type UserData struct {
	Name     *string `json:"name"`
	Birthday *string `json:"birthday"`
	Age      *int
	IsMale   *bool `json:"is_male"`
}

func badResponse(w http.ResponseWriter, code int, message string) {
	response(w, code, map[string]string{"error": message})
}

func response(w http.ResponseWriter, code int, payload interface{}) {
	body, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(body)
}

func (a *App) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := UsersGetAll(r.Context(), a.db)
	if err != nil {
		badResponse(w, http.StatusInternalServerError, "Something going wrong")
		return
	}
	response(w, http.StatusOK, users)
}

func parseMingAge(r *http.Request) (int, error) {
	keys, ok := r.URL.Query()["minAge"]
	if !ok || len(keys) != 1 {
		return 0, fmt.Errorf("cann't parse minAge")
	}
	minAge, err := strconv.ParseInt(keys[0], 10, 64)
	return int(minAge), err
}

func (a *App) GetUsersWithMinAge(w http.ResponseWriter, r *http.Request) {
	minAge, err := parseMingAge(r)
	if err != nil {
		log.Println(err)
		badResponse(w, http.StatusBadRequest, "minAge must be number")
		return
	}
	users, err := UsersGetWithMinAge(r.Context(), a.db, minAge)
	if err != nil {
		badResponse(w, http.StatusInternalServerError, "internal error")
		return
	}
	response(w, http.StatusOK, users)
}

func parseId(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return int64(id), nil
}

func (a *App) GetUserById(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r)
	if err != nil {
		log.Println(err)
		badResponse(w, http.StatusBadRequest, "incorect id")
		return
	}
	user, err := UsersGetById(r.Context(), a.db, id)
	if err != nil {
		if errors.Is(err, NotFoundError) {
			badResponse(w, http.StatusNotFound, "not found")
		} else {
			log.Println(err)
			badResponse(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	response(w, http.StatusOK, user)
}

func getAge(date time.Time) int {
	now := time.Now()
	age := now.Year() - date.Year()
	if now.Month() > date.Month() {
		age++
	} else if date.Month() == now.Month() {
		if date.Day() < now.Day() {
			age++
		} else if date.Day() > now.Day() {
			age--
		}
	} else {
		age--
	}
	return age
}

func (a *App) GreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var user UserData
	if err := json.Unmarshal(body, &user); err != nil {
		log.Println(err)
		badResponse(w, http.StatusBadRequest, "bad request")
		return
	}

	if user.Birthday != nil {
		layout := "02.01.2006"
		date, err := time.Parse(layout, *user.Birthday)
		if err != nil {
			log.Println(err)
			badResponse(w, http.StatusBadRequest, "date is incorect")
			return
		}
		age := getAge(date)
		user.Age = &age
	}

	if id, err := UsersInsert(r.Context(), a.db, user); err != nil {
		log.Println(err)
		badResponse(w, http.StatusBadRequest, "bad request")
	} else {
		response(w, http.StatusOK, map[string]int64{"id": id})
	}
}

func (a *App) ChangeUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r)
	if err != nil {
		log.Println(err)
		badResponse(w, http.StatusBadRequest, "incorect id")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		log.Println(err)
		badResponse(w, http.StatusBadRequest, "bad request")
		return
	}
	user.Id = id

	if user.Birthday != nil {
		layout := "02.01.2006"
		date, err := time.Parse(layout, *user.Birthday)
		if err != nil {
			log.Println(err)
			badResponse(w, http.StatusBadRequest, "date is incorect")
			return
		}
		age := getAge(date)
		user.Age = &age
	}

	if err := UsersUpdateById(r.Context(), a.db, user); err != nil {
		if errors.Is(err, NotUpdatedError) {
			badResponse(w, http.StatusNotFound, "not found")
		} else {
			log.Println(err)
			badResponse(w, http.StatusBadRequest, "bad request")
		}
		return
	}
	response(w, http.StatusOK, "")
}

func (a *App) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r)
	if err != nil {
		log.Println(err)
		badResponse(w, http.StatusBadRequest, "incorect id")
		return
	}

	if err := UsersDeleteById(r.Context(), a.db, id); err != nil {
		if errors.Is(err, NotFoundError) {
			badResponse(w, http.StatusNotFound, "not found")
		} else {
			log.Println(err)
			badResponse(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	response(w, http.StatusOK, "")
}
