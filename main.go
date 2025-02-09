package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

type Book struct {
	ID     int       `json:"id"`
	Isbn   string    `json:"isbn"`
	Title  string    `json:"title"`
	Author *[]Author `json:"author"`
}

type Author struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

var books []Book

func getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sort.Slice(books, func(i, j int) bool {
		return books[i].ID < books[j].ID
	})
	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, item := range books {
		idParam, err := strconv.Atoi(params["id"])
		if err != nil {
			log.Println("Invalid ID format")
			return
		}

		if item.ID == idParam {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Book{})
}

func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book Book

	// ตรวจสอบข้อผิดพลาดในการ Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// ตรวจสอบว่า books ว่างหรือไม่ และหาค่า ID สูงสุด
	maxID := 0
	for _, b := range books {
		if b.ID > maxID {
			maxID = b.ID
		}
	}
	book.ID = maxID + 1

	books = append(books, book)
	json.NewEncoder(w).Encode(book)
}

// update
func updateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	idParam, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	for index, item := range books {
		if item.ID == idParam {
			// ลบหนังสือเก่าที่ตรงกับ ID
			books = append(books[:index], books[index+1:]...)
			// แปลงข้อมูลจาก Body เป็น Book ใหม่
			var book Book
			if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
				http.Error(w, "Invalid book data", http.StatusBadRequest)
				return
			}
			// make not change sorting

			// book.ID = idParam
			books = append(books, book)

			json.NewEncoder(w).Encode(book)
			return
		}
	}

	var buffer bytes.Buffer

	if err := json.NewEncoder(&buffer).Encode(books); err != nil {
		http.Error(w, "Failed to encode books", http.StatusInternalServerError)
		return
	}

	errorMsg := "Book not found: " + buffer.String()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	http.Error(w, errorMsg, http.StatusNotFound)
}

// delete
func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range books {
		idParam, err := strconv.Atoi(params["id"])
		if err != nil {
			log.Println("Invalid ID format")
			return
		}

		if item.ID == idParam {
			// delete item
			books = append(books[:index], books[index+1:]...)

			json.NewEncoder(w).Encode(item)
			return
		}
	}

	var buffer bytes.Buffer

	if err := json.NewEncoder(&buffer).Encode(books); err != nil {
		http.Error(w, "Failed to encode books", http.StatusInternalServerError)
		return
	}

	errorMsg := "Book not found" + buffer.String()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	http.Error(w, errorMsg, http.StatusNotFound)
}

func main() {
	r := mux.NewRouter()
	// append another book
	values := []string{"albert", "berina", "charles"}
	authors := []Author{}

	for _, val := range values {
		author := Author{
			Firstname: fmt.Sprintf("John %s", val),
			Lastname:  fmt.Sprintf("Doe %s", val),
		}
		authors = append(authors, author)
	}
	books = append(books, Book{ID: 3, Isbn: "12345", Title: "Book 3", Author: &authors})
	books = append(books, Book{ID: 3, Isbn: "67890", Title: "Book 3", Author: &[]Author{
		{"oak", "kasidit"},
		{Firstname: "John", Lastname: "Smith"},
		{Firstname: "Jane", Lastname: "Doe"},
		{values[0], values[1]},
	}})

	r.HandleFunc("/books", getBooks).Methods("GET")
	r.HandleFunc("/book/{id}", getBook).Methods("GET")
	r.HandleFunc("/book", createBook).Methods("POST")
	// update
	r.HandleFunc("/book/{id}", updateBook).Methods("PUT")
	// delete
	r.HandleFunc("/book/{id}", deleteBook).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}
