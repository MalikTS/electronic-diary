// db/queries.go
package db

import (
	"context"
	"electronic-diary/models"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetGroupIDByName(name string) (primitive.ObjectID, error) {
	ctx := context.Background()
	var group models.Group
	err := groupsCol.FindOne(ctx, bson.M{"name": name}).Decode(&group)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return group.ID, nil
}

func GetStudentsByGroupID(groupID primitive.ObjectID) ([]models.Student, error) {
	ctx := context.Background()
	cursor, err := studentsCol.Find(ctx, bson.M{"groupId": groupID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var students []models.Student
	err = cursor.All(ctx, &students)
	return students, err
}

func GetStudentByID(id primitive.ObjectID) (*models.Student, error) {
	ctx := context.Background()
	var student models.Student
	err := studentsCol.FindOne(ctx, bson.M{"_id": id}).Decode(&student)
	return &student, err
}

func GetDisciplinesByGroupID(groupID primitive.ObjectID) ([]models.Discipline, error) {
	ctx := context.Background()
	cursor, err := disciplinesCol.Find(ctx, bson.M{"groupId": groupID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var disciplines []models.Discipline
	err = cursor.All(ctx, &disciplines)
	return disciplines, err
}

func GetStudentDisciplineData(studentID primitive.ObjectID) ([]models.StudentDisciplineData, error) {
	ctx := context.Background()
	cursor, err := studentDisciplineDataCol.Find(ctx, bson.M{"studentId": studentID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var data []models.StudentDisciplineData
	err = cursor.All(ctx, &data)
	return data, err
}

func UpdateStudent(studentID primitive.ObjectID, comments string) error {
	ctx := context.Background()
	_, err := studentsCol.UpdateOne(ctx, bson.M{"_id": studentID}, bson.M{"$set": bson.M{"comments": comments}})
	return err
}

func UpdateDisciplineData(dataID primitive.ObjectID, score, total, attended int) error {
	ctx := context.Background()
	_, err := studentDisciplineDataCol.UpdateOne(ctx,
		bson.M{"_id": dataID},
		bson.M{"$set": bson.M{
			"score":            score,
			"totalClasses":     total,
			"attendedClasses":  attended,
		}},
	)
	return err
}

func ResetDynamicData() error {
	ctx := context.Background()

	// Обнуляем комментарии у всех студентов
	_, err := studentsCol.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"comments": ""}})
	if err != nil {
		return err
	}

	// Обнуляем данные по дисциплинам
	_, err = studentDisciplineDataCol.UpdateMany(ctx, bson.M{}, bson.M{
		"$set": bson.M{
			"score":            0,
			"totalClasses":     0,
			"attendedClasses":  0,
		},
	})
	if err != nil {
		return err
	}

	return nil
}