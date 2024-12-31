package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type Student struct {
	Name  string `json:"name"`
	Class string `json:"class"`
	Age   int    `json:"age"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./student.db") 
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("error pinging database:", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		class TEXT,
		age INTEGER
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal("error creating table:", err)
	}
}

func initHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Student API - Welcome to %s!", r.URL.Path[1:])
}

func getAllStudents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("SELECT name, class, age FROM students")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var student Student
		if err := rows.Scan(&student.Name, &student.Class, &student.Age); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		students = append(students, student)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(students)
}

func addStudents(w http.ResponseWriter, r *http.Request) {
	var student Student
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO students (name, class, age) VALUES (?, ?, ?)", student.Name, student.Class, student.Age)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

func handleRequests() {
	http.HandleFunc("/", initHandler)
	http.HandleFunc("/students", getAllStudents)
	http.HandleFunc("/addStudents", addStudents)
}

func main() {
	initDB()
	defer db.Close() // Ensure the database connection is closed on exit

	// Insert initial data
	_, _ = db.Exec("INSERT INTO students (name, class, age) VALUES (?, ?, ?)", "John Audu", "SS1", 15)
	_, _ = db.Exec("INSERT INTO students (name, class, age) VALUES (?, ?, ?)", "Kelly Mba", "SS2", 17)

	handleRequests()
	fmt.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
