// main.go
package main

import (
	"electronic-diary/db"
	"electronic-diary/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	// Проверка --reset
	reset := false
	for _, arg := range os.Args {
		if arg == "--reset" {
			reset = true
			break
		}
	}

	db.Connect()
	db.InitCollections()

	if reset {
		db.ResetData()
	}
	db.SeedData()

	// Регистрируем маршруты
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/group/", handlers.GroupHandler)
	http.HandleFunc("/student/", handlers.StudentHandler)
	http.HandleFunc("/api/student/", handlers.UpdateStudentHandler)
	http.HandleFunc("/api/reset-dynamic", handlers.ResetDynamicHandler)

	// Статические файлы (CSS/JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	log.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))	
}