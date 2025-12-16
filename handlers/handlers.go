// handlers/handlers.go
package handlers

import (
	"context"
	"electronic-diary/db"
	"electronic-diary/models"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Утилита: преобразовать строку в ObjectID
func parseObjectID(s string) (primitive.ObjectID, error) {
	if !primitive.IsValidObjectID(s) {
		return primitive.NilObjectID, fmt.Errorf("invalid ID")
	}
	return primitive.ObjectIDFromHex(s)
}

// Главная страница
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<title>Электронный дневник</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<div class="card">
		<div class="home-header">
			<h1>Электронный дневник</h1>
		</div>
		<div class="home-buttons">
			<a href="/group/Backend" class="main-group-btn"><button>Backend</button></a>
			<a href="/group/Frontend" class="main-group-btn"><button>Frontend</button></a>
		</div>
		<div class="home-reset">
			<form action="/api/reset-dynamic" method="POST" onsubmit="return confirm('Обнулить все баллы, посещаемость и комментарии? Это нельзя отменить!')">
				<button type="submit" class="reset-btn">Сбросить динамические данные</button>
			</form>
		</div>
	</div>
</body>
</html>`
	w.Write([]byte(tmpl))
}

// Страница группы — показывает студентов
func GroupHandler(w http.ResponseWriter, r *http.Request) {
	groupName := strings.TrimPrefix(r.URL.Path, "/group/")
	if groupName != "Backend" && groupName != "Frontend" {
		http.NotFound(w, r)
		return
	}

	groupID, err := db.GetGroupIDByName(groupName)
	if err != nil {
		http.Error(w, "Группа не найдена", http.StatusNotFound)
		return
	}

	students, err := db.GetStudentsByGroupID(groupID)
	if err != nil {
		log.Printf("Ошибка получения студентов: %v", err)
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<title>Группа {{.GroupName}}</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<div class="card">
		<h1>Группа: {{.GroupName}}</h1>
		<div class="group-student-list">
		{{range .Students}}
			<a href="/student/{{.ID.Hex}}">{{.Name}}</a>
		{{end}}
		</div>
		<a href="/" class="back-link">← Назад</a>
	</div>
</body>
</html>`

	data := struct {
		GroupName string
		Students  []models.Student
	}{
		GroupName: groupName,
		Students:  students,
	}

	t := template.Must(template.New("group").Parse(tmpl))
	t.Execute(w, data)
}

// Страница студента — покажем детали
func StudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/student/")
	studentID, err := parseObjectID(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	student, err := db.GetStudentByID(studentID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	groupID := student.GroupID
	disciplines, err := db.GetDisciplinesByGroupID(groupID)
	if err != nil {
		http.Error(w, "Ошибка дисциплин", http.StatusInternalServerError)
		return
	}

	disciplineData, err := db.GetStudentDisciplineData(studentID)
	if err != nil {
		http.Error(w, "Ошибка данных", http.StatusInternalServerError)
		return
	}

	dataMap := make(map[primitive.ObjectID]models.StudentDisciplineData)
	for _, d := range disciplineData {
		dataMap[d.DisciplineID] = d
	}

	// Подготовим данные для статистики
	var bestScore, worstScore *models.StudentDisciplineData
	var bestAttendance, worstAttendance *models.StudentDisciplineData
	var allData []models.StudentDisciplineData
	for _, d := range dataMap {
		allData = append(allData, d)
	}

	if len(allData) > 0 {
		bestScore = &allData[0]
		worstScore = &allData[0]
		bestAttendance = &allData[0]
		worstAttendance = &allData[0]

		for _, d := range allData {
			if d.Score > bestScore.Score {
				bestScore = &d
			}
			if d.Score < worstScore.Score {
				worstScore = &d
			}
			attPerc := 0.0
			if d.TotalClasses > 0 {
				attPerc = float64(d.AttendedClasses) / float64(d.TotalClasses) * 100
			}
			bestAttPerc := 0.0
			if bestAttendance.TotalClasses > 0 {
				bestAttPerc = float64(bestAttendance.AttendedClasses) / float64(bestAttendance.TotalClasses) * 100
			}
			worstAttPerc := 0.0
			if worstAttendance.TotalClasses > 0 {
				worstAttPerc = float64(worstAttendance.AttendedClasses) / float64(worstAttendance.TotalClasses) * 100
			}

			if attPerc > bestAttPerc {
				bestAttendance = &d
			}
			if attPerc < worstAttPerc {
				worstAttendance = &d
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Student.Name}}</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<div class="card">
		<h1>{{.Student.Name}}</h1>

		<form id="save-form" method="POST" action="/api/student/{{.Student.ID.Hex}}">
	<div class="comment-area">
		<label for="comments">Комментарий:</label>
		<textarea name="comments" id="comments" placeholder="Введите комментарий...">{{.Student.Comments}}</textarea>
	</div>

	<h2>Дисциплины</h2>
	<table id="disciplines-table">
		<thead>
			<tr>
				<th>Дисциплина</th>
				<th>Баллы (0–100)</th>
				<th>Всего пар</th>
				<th>Посетил</th>
				<th>%</th>
				<th>Оценка</th>
			</tr>
		</thead>
		<tbody>
		{{range .Disciplines}}
		{{$data := index $.DataMap .ID}}
		<tr data-disc-id="{{.ID.Hex}}">
			<td>{{.Name}}</td>
			<td><input type="number" name="score_{{.ID.Hex}}" class="score-input" data-disc="{{.ID.Hex}}" value="{{$data.Score}}" min="0" max="100"></td>
			<td><input type="number" name="total_{{.ID.Hex}}" class="total-input" data-disc="{{.ID.Hex}}" value="{{$data.TotalClasses}}" min="0"></td>
			<td><input type="number" name="attended_{{.ID.Hex}}" class="attended-input" data-disc="{{.ID.Hex}}" value="{{$data.AttendedClasses}}" min="0"></td>
			<td class="perc-cell">
				{{if gt $data.TotalClasses 0}}
					{{printf "%.0f" (div (mul $data.AttendedClasses 100) $data.TotalClasses)}}
				{{else}}
					0
				{{end}}%
			</td>
			<td class="grade-cell {{gradeClass $data.Score}}">{{scoreToGrade $data.Score}}</td>
		</tr>
		{{end}}
		</tbody>
	</table>

	<div class="statistics">
		<h3>Статистика</h3>
		{{if $.BestScore}}
		<div class="stat-item">Лучший предмет: <strong>{{getDiscName $.Disciplines $.BestScore.DisciplineID}} ({{$.BestScore.Score}} баллов)</strong></div>
		<div class="stat-item">Слабый предмет: <strong>{{getDiscName $.Disciplines $.WorstScore.DisciplineID}} ({{$.WorstScore.Score}} баллов)</strong></div>
		<div class="stat-item">Лучшая посещаемость: <strong>{{getDiscName $.Disciplines $.BestAttendance.DisciplineID}} ({{calcPerc $.BestAttendance.AttendedClasses $.BestAttendance.TotalClasses}}%)</strong></div>
		<div class="stat-item">Худшая посещаемость: <strong>{{getDiscName $.Disciplines $.WorstAttendance.DisciplineID}} ({{calcPerc $.WorstAttendance.AttendedClasses $.WorstAttendance.TotalClasses}}%)</strong></div>
		{{end}}
	</div>

	<input type="submit" value="Сохранить">
</form>

<a href="/group/{{.GroupName}}" class="back-link">← Назад к группе</a>
	</div>

	<script>
		function scoreToGrade(score) {
			score = parseInt(score);
			if (score >= 80) return 5;
			if (score >= 60) return 4;
			if (score >= 40) return 3;
			if (score >= 20) return 2;
			return 1;
		}

		function gradeClass(score) {
			const grade = scoreToGrade(score);
			return 'grade-' + grade;
		}

		function updateRow(discId) {
			const row = document.querySelector('tr[data-disc-id="' + discId + '"]');
			const scoreInput = row.querySelector('.score-input');
			const totalInput = row.querySelector('.total-input');
		 const attendedInput = row.querySelector('.attended-input');
			const percCell = row.querySelector('.perc-cell');
			const gradeCell = row.querySelector('.grade-cell');

			const total = parseInt(totalInput.value) || 0;
			const attended = parseInt(attendedInput.value) || 0;
			const score = parseInt(scoreInput.value) || 0;

			// Валидация
			let isValid = true;
			if (score < 0 || score > 100) isValid = false;
			if (attended > total) isValid = false;

			scoreInput.classList.toggle('error', score < 0 || score > 100);
			attendedInput.classList.toggle('error', attended > total);
			totalInput.classList.toggle('error', attended > total);

			// Обновляем % и оценку
			let perc = total > 0 ? Math.round((attended / total) * 100) : 0;
			percCell.textContent = perc + '%';
			gradeCell.textContent = scoreToGrade(score);
			gradeCell.className = 'grade-cell ' + gradeClass(score);

			return isValid;
		}

		// Навешиваем слушатели
		document.querySelectorAll('.score-input, .total-input, .attended-input').forEach(input => {
			input.addEventListener('input', function() {
				const discId = this.dataset.disc;
				updateRow(discId);
			});
		});

		// Форма сохранения
		document.getElementById('save-form').addEventListener('submit', function(e) {
			const comment = document.getElementById('comments').value;
			document.getElementById('hidden-comments').value = comment;

			// Проверка валидации
			let allValid = true;
			document.querySelectorAll('.score-input, .attended-input, .total-input').forEach(input => {
				if (input.classList.contains('error')) allValid = false;
			});

			if (!allValid) {
				e.preventDefault();
				alert('Исправьте ошибки в данных (баллы 0–100, посещаемость ≤ общего числа пар).');
				return;
			}
		});
	</script>
</body>
</html>`

	funcMap := template.FuncMap{
		"div": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"scoreToGrade": func(score int) int {
			switch {
			case score >= 80: return 5
			case score >= 60: return 4
			case score >= 40: return 3
			case score >= 20: return 2
			default: return 1
			}
		},
		"gradeClass": func(score int) string {
			switch {
			case score >= 80: return "grade-5"
			case score >= 60: return "grade-4"
			case score >= 40: return "grade-3"
			case score >= 20: return "grade-2"
			default: return "grade-1"
			}
		},
		"calcPerc": func(attended, total int) int {
			if total == 0 {
				return 0
			}
			return int(float64(attended) / float64(total) * 100)
		},
		"getDiscName": func(disciplines []models.Discipline, id primitive.ObjectID) string {
			for _, d := range disciplines {
				if d.ID == id {
					return d.Name
				}
			}
			return "—"
		},
	}

	t := template.Must(template.New("student").Funcs(funcMap).Parse(tmpl))

	groupName := "Backend"
	if groupID == (func() primitive.ObjectID {
		id, _ := db.GetGroupIDByName("Frontend")
		return id
	}()) {
		groupName = "Frontend"
	}

	data := struct {
		Student           *models.Student
		Disciplines       []models.Discipline
		DataMap           map[primitive.ObjectID]models.StudentDisciplineData
		GroupName         string
		BestScore         *models.StudentDisciplineData
		WorstScore        *models.StudentDisciplineData
		BestAttendance    *models.StudentDisciplineData
		WorstAttendance   *models.StudentDisciplineData
	}{
		Student:           student,
		Disciplines:       disciplines,
		DataMap:           dataMap,
		GroupName:         groupName,
		BestScore:         bestScore,
		WorstScore:        worstScore,
		BestAttendance:    bestAttendance,
		WorstAttendance:   worstAttendance,
	}

	t.Execute(w, data)
}

// Обработка сохранения
func UpdateStudentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/student/")
	studentID, err := parseObjectID(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Обновляем комментарий
	comments := r.FormValue("comments")
	_ = db.UpdateStudent(studentID, comments)

	r.ParseForm()
	ctx := context.Background()

	for key, values := range r.Form {
		if strings.HasPrefix(key, "score_") {
			discIDHex := strings.TrimPrefix(key, "score_")
			discID, err := parseObjectID(discIDHex)
			if err != nil {
				log.Printf("Некорректный ID дисциплины: %s", discIDHex)
				continue
			}

			score, _ := strconv.Atoi(values[0])
			total, _ := strconv.Atoi(r.FormValue("total_" + discIDHex))
			attended, _ := strconv.Atoi(r.FormValue("attended_" + discIDHex))

			// Пытаемся найти существующую запись
			var data models.StudentDisciplineData
			err = db.StudentDisciplineDataCol().FindOne(ctx, bson.M{
				"studentId":    studentID,
				"disciplineId": discID,
			}).Decode(&data)

			if err == nil {
				// Запись найдена — обновляем
				_ = db.UpdateDisciplineData(data.ID, score, total, attended)
			} else if err == mongo.ErrNoDocuments {
				// Записи нет — создаём новую
				newData := models.StudentDisciplineData{
					StudentID:       studentID,
					DisciplineID:    discID,
					Score:           score,
					TotalClasses:    total,
					AttendedClasses: attended,
				}
				_, insertErr := db.StudentDisciplineDataCol().InsertOne(ctx, newData)
				if insertErr != nil {
					log.Printf("Ошибка при создании записи для студента %s, дисциплины %s: %v", idStr, discIDHex, insertErr)
				}
			} else {
				// Другая ошибка (например, БД недоступна)
				log.Printf("Ошибка при поиске записи: %v", err)
			}
		}
	}

	// Перенаправляем на страницу студента — данные загрузятся свежие из БД
	http.Redirect(w, r, "/student/"+idStr, http.StatusSeeOther)
}	


func ResetDynamicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	err := db.ResetDynamicData()
	if err != nil {
		log.Printf("Ошибка сброса: %v", err)
		http.Error(w, "Ошибка при сбросе", http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON-ответ (для JS) или редирект
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
