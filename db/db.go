// db/db.go
package db

import (
	"context"
	"log"
	"time"
	
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func StudentDisciplineDataCol() *mongo.Collection {
	return studentDisciplineDataCol
}

func Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Не удалось подключиться к MongoDB:", err)
	}

	// Проверим подключение
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Не удалось пингануть MongoDB:", err)
	}

	DB = client.Database("electronic_diary")
	log.Println("✅ Подключились к MongoDB: electronic_diary")
}

