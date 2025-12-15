// models/models.go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Group struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name"`
}

type Student struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	GroupID  primitive.ObjectID `bson:"groupId" json:"groupId"`
	Comments string             `bson:"comments" json:"comments"`
}

type Discipline struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `bson:"name" json:"name"`
	GroupID primitive.ObjectID `bson:"groupId" json:"groupId"`
}

type StudentDisciplineData struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID        primitive.ObjectID `bson:"studentId" json:"studentId"`
	DisciplineID     primitive.ObjectID `bson:"disciplineId" json:"disciplineId"`
	Score            int                `bson:"score" json:"score"`
	TotalClasses     int                `bson:"totalClasses" json:"totalClasses"`
	AttendedClasses  int                `bson:"attendedClasses" json:"attendedClasses"`
}