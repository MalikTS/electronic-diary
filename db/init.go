// db/init.go
package db

import (
	"context"
	"log"

	"electronic-diary/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	groupsCol               *mongo.Collection
	studentsCol             *mongo.Collection
	disciplinesCol          *mongo.Collection
	studentDisciplineDataCol *mongo.Collection
)

func InitCollections() {
	groupsCol = DB.Collection("groups")
	studentsCol = DB.Collection("students")
	disciplinesCol = DB.Collection("disciplines")
	studentDisciplineDataCol = DB.Collection("studentDisciplineData")
}

func SeedData() {
	ctx := context.Background()

	// Проверим, есть ли уже группы
	count, _ := groupsCol.CountDocuments(ctx, bson.M{})
	if count > 0 {
		log.Println("📚 Данные уже существуют — пропускаем инициализацию.")
		return
	}

	log.Println("🌱 Инициализация начальных данных...")

	// === 1. Создаём группы ===
	backendGroup := models.Group{Name: "Backend"}
	frontendGroup := models.Group{Name: "Frontend"}

	backendResult, _ := groupsCol.InsertOne(ctx, backendGroup)
	frontendResult, _ := groupsCol.InsertOne(ctx, frontendGroup)

	backendID := backendResult.InsertedID.(primitive.ObjectID)
	frontendID := frontendResult.InsertedID.(primitive.ObjectID)

	// === 2. Создаём студентов ===
	// Backend студенты
backendNames := []string{
    "Магомед Магомедов", "Хамхоев Иса", "Мархиев Ислам",
}
// Frontend студенты
frontendNames := []string{
	"Костоева Залина", "Цечоев Абдула", "Татиев Илез", "Татиев Хамзат", "Чиниев Ильяс", "Точиев Рамзан",
}

var students []interface{}
for _, name := range backendNames {
    students = append(students, models.Student{Name: name, GroupID: backendID, Comments: ""})
}
for _, name := range frontendNames {
    students = append(students, models.Student{Name: name, GroupID: frontendID, Comments: ""})
}
	studentsCol.InsertMany(ctx, students)

	// === 3. Создаём дисциплины ===
	backendDisciplines := []string{
		"GO",
		"Node.js",
		"Основы Linux",
		"Алгоритмы и структуры данных",
		"Английский язык",
	}
	frontendDisciplines := []string{
		"Английский язык",
		"JavaScript Framework",
		"HTML5",
		"CSS",
		"Web-компоненты",
	}

	var disciplines []interface{}
	for _, name := range backendDisciplines {
		disciplines = append(disciplines, models.Discipline{Name: name, GroupID: backendID})
	}
	for _, name := range frontendDisciplines {
		disciplines = append(disciplines, models.Discipline{Name: name, GroupID: frontendID})
	}
	disciplinesCol.InsertMany(ctx, disciplines)

	// === 4. Получим всех студентов и дисциплины для связи ===
	var allStudents []models.Student
	var allDisciplines []models.Discipline

	cursor, err := studentsCol.Find(ctx, bson.M{})
		if err != nil {
    log.Fatal(err)
		}
		defer cursor.Close(ctx)
		cursor.All(ctx, &allStudents)

	cursor1, err1 := disciplinesCol.Find(ctx, bson.M{})
		if err1 != nil {
    log.Fatal(err1)
		}
		defer cursor1.Close(ctx)
		cursor1.All(ctx, &allDisciplines)

	// === 5. Создаём StudentDisciplineData ===
	var dataEntries []interface{}
	for _, student := range allStudents {
		for _, disc := range allDisciplines {
			if disc.GroupID == student.GroupID {
				dataEntries = append(dataEntries, models.StudentDisciplineData{
					StudentID:       student.ID,
					DisciplineID:    disc.ID,
					Score:           0,
					TotalClasses:    0,
					AttendedClasses: 0,
				})
			}
		}
	}
	studentDisciplineDataCol.InsertMany(ctx, dataEntries)

	log.Println("✅ Начальные данные успешно созданы!")

	
}

func ResetData() {
	ctx := context.Background()
	
	// Удаляем все коллекции
	groupsCol.Drop(ctx)
	studentsCol.Drop(ctx)
	disciplinesCol.Drop(ctx)
	studentDisciplineDataCol.Drop(ctx)

	log.Println("Все данные удалены.")
}