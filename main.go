package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

var tpl *template.Template
var db *sql.DB

func main() {
	tpl, _ = template.ParseGlob("templates/*.html")
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(localhost:3306)/demogo")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	http.HandleFunc("/insert", insertHandler)
	http.HandleFunc("/browse", browseHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/updateresult", updateResultHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/", homePageHandler)
	http.ListenAndServe("localhost:8080", nil)
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****browseHandler running*****")
	stmt := "SELECT * FROM users"
	rows, err := db.Query(stmt)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var users []User

	for rows.Next() {
		var u User
		err = rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password)
		if err != nil {
			panic(err)
		}
		users = append(users, u)
	}
	tpl.ExecuteTemplate(w, "select.html", users)
}

func insertHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tpl.ExecuteTemplate(w, "insert.html", nil)
	case "POST":
		r.ParseForm()
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		var err error
		if name == "" || email == "" || password == "" {
			fmt.Println("Error inserting row:", err)
			tpl.ExecuteTemplate(w, "insert.html", "Error inserting data, please check all fields.")
			return
		}
		ins, err := db.Prepare("INSERT INTO `demogo`.`users` (`name`, `email`, `password`) VALUES (?, ?, ?);")
		if err != nil {
			panic(err)
		}
		defer ins.Close()
		_, err = ins.Exec(name, email, password)
		if err != nil {
			fmt.Println("Error inserting row:", err)
			tpl.ExecuteTemplate(w, "insert.html", "Error inserting data, please check all fields.")
			return
		}
		tpl.ExecuteTemplate(w, "insert.html", "User Successfully Inserted")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		r.ParseForm()
		id := r.FormValue("id")
		row := db.QueryRow("SELECT * FROM users WHERE id = ?;", id)
		var u User
		err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password)
		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/browse", http.StatusSeeOther)
			return
		}
		tpl.ExecuteTemplate(w, "update.html", u)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func updateResultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		r.ParseForm()
		id := r.FormValue("id")
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		upStmt := "UPDATE `demogo`.`users` SET `name` = ?, `email` = ?, `password` = ? WHERE (`id` = ?);"
		stmt, err := db.Prepare(upStmt)
		if err != nil {
			fmt.Println("Error preparing statement:", err)
			panic(err)
		}
		defer stmt.Close()
		_, err = stmt.Exec(name, email, password, id)
		if err != nil {
			fmt.Println("Error updating row:", err)
			tpl.ExecuteTemplate(w, "result.html", "There was a problem updating the user")
			return
		}
		tpl.ExecuteTemplate(w, "result.html", "User was Successfully Updated")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		r.ParseForm()
		id := r.FormValue("id")
		del, err := db.Prepare("DELETE FROM `demogo`.`users` WHERE (`id` = ?);")
		if err != nil {
			panic(err)
		}
		defer del.Close()
		_, err = del.Exec(id)
		if err != nil {
			fmt.Fprint(w, "Error deleting user")
			return
		}
		tpl.ExecuteTemplate(w, "result.html", "User was Successfully Deleted")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/browse", http.StatusSeeOther)
}
