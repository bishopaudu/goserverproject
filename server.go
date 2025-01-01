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
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone_number"`
	Nationality   string `json:"nationality"`
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
		age INTEGER,
		email TEXT,
		phone_number TEXT,
		nationality TEXT
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal("error creating table:", err)
	}

	insertInitialData()
}

func insertInitialData() {
	initialData := []Student{
		{"Alice Johnson", "Grade 10", 15, "alice@example.com", "1234567890", "American"},
		{"Bob Smith", "Grade 12", 17, "bob@example.com", "0987654321", "British"},
		{"Charlie Brown", "Grade 11", 16, "charlie@example.com", "1122334455", "Canadian"},
	}

	for _, student := range initialData {
		_, err := db.Exec(
			"INSERT INTO students (name, class, age, email, phone_number, nationality) VALUES (?, ?, ?, ?, ?, ?)",
			student.Name, student.Class, student.Age, student.Email, student.PhoneNumber, student.Nationality,
		)
		if err != nil {
			log.Printf("Error inserting initial data for %s: %v", student.Name, err)
		}
	}
}


func initHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Student API - Welcome to %s!", r.URL.Path[1:])
}

func getAllStudents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("SELECT name, class, age, email,phone_number,nationality FROM students")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var student Student
		if err := rows.Scan(&student.Name, &student.Class, &student.Age, &student.Email, &student.PhoneNumber, &student.Nationality); err != nil {
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

	// Decode JSON payload
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Log the decoded student
	log.Printf("Decoded student: %+v", student)

	// SQL INSERT query
	query := "INSERT INTO students (name, class, age, email, phone_number, nationality) VALUES (?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, student.Name, student.Class, student.Age, student.Email, student.PhoneNumber, student.Nationality)
	if err != nil {
		log.Printf("Error executing query: %s | Error: %v", query, err)
		http.Error(w, "Error inserting into database", http.StatusInternalServerError)
		return
	}

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error fetching RowsAffected: %v", err)
		http.Error(w, "Error fetching rows affected", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Println("No rows were inserted")
		http.Error(w, "No rows inserted", http.StatusInternalServerError)
		return
	}

	// Log success and respond
	log.Println("Successfully inserted student")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}
func searchStudents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Retrieve query parameters
	name := r.URL.Query().Get("name")
	nationality := r.URL.Query().Get("nationality")

	// Build SQL query
	query := "SELECT name, class, age, email, phone_number, nationality FROM students WHERE 1=1"
	args := []interface{}{}

	if name != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+name+"%")
	}

	if nationality != "" {
		query += " AND nationality LIKE ?"
		args = append(args, "%"+nationality+"%")
	}

	// Log the query for debugging
	log.Printf("Executing query: %s with args: %v", query, args)

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Query error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	// Parse results
	var students []Student
	for rows.Next() {
		var student Student
		if err := rows.Scan(&student.Name, &student.Class, &student.Age, &student.Email, &student.PhoneNumber, &student.Nationality); err != nil {
			log.Printf("Scan error: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		students = append(students, student)
	}

	// Check for iteration errors
	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode and send response
	json.NewEncoder(w).Encode(students)
}


func handleRequests() {
	http.HandleFunc("/", initHandler)
	http.HandleFunc("/students", getAllStudents)
	http.HandleFunc("/addStudents", addStudents)
	http.HandleFunc("/search", searchStudents)
}

func main() {
	initDB()
	defer db.Close() // Ensure the database connection is closed on exit
	handleRequests()
	fmt.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
