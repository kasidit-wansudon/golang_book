package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

type Book struct {
	ID     int     `json:"id"`
	Isbn   string  `json:"isbn"`
	Title  string  `json:"title"`
	Author *Author `json:"author"`
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

// create
func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book Book
	_ = json.NewDecoder(r.Body).Decode(&book)
	book.ID = len(books) + 1
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

			book.ID = idParam
			books = append(books, book)
			// sort by id
			for i := 0; i < len(books)-1; i++ {
				if books[i].ID > books[i+1].ID {
					books[i], books[i+1] = books[i+1], books[i]
				}
			}

			json.NewEncoder(w).Encode(book)
			return
		}
	}

	http.Error(w, "Book not found", http.StatusNotFound)
}

// delete
func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, item := range books {
		idParam, err := strconv.Atoi(params["id"])
		if err != nil {
			log.Println("Invalid ID format")
			return
		}

		if item.ID == idParam {
			// delete
			books = append(books[:idParam-1], books[idParam:]...)

			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(books)
}

func main() {
	r := mux.NewRouter()
	books = []Book{(Book{ID: 1, Isbn: "12345", Title: "Book 1", Author: &Author{Firstname: "John", Lastname: "Doe"}})}
	// append another book
	books = append(books, Book{ID: 2, Isbn: "54321", Title: "Book 2", Author: &Author{Firstname: "Jane", Lastname: "Doe"}})
	// append another book
	books = append(books, Book{ID: 3, Isbn: "12345", Title: "Book 3", Author: &Author{Firstname: "John", Lastname: "Doe"}})

	r.HandleFunc("/books", getBooks).Methods("GET")
	r.HandleFunc("/book/{id}", getBook).Methods("GET")
	r.HandleFunc("/book", createBook).Methods("POST")
	// update
	r.HandleFunc("/book/{id}", updateBook).Methods("PUT")
	// delete
	r.HandleFunc("/book/{id}", deleteBook).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}
