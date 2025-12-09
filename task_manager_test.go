package main

import (
	"encoding/csv"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testFilename = "test_tasks.json"
const testCSVFilename = "test_export.csv"

func setupTestManager() *TaskManager {
	os.Remove(testFilename)    // Удаляем файл, если он существует
	os.Remove(testCSVFilename) // Удаляем файл экспорта, если он существует
	return NewTaskManager(testFilename)
}

func teardownTestManager() {
	os.Remove(testFilename)
	os.Remove(testCSVFilename)
}

func TestAddTask(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	// Добавляем задачу
	title := "Test Task"
	description := "Test Description"
	priority := 2
	dueDate := time.Now().Add(24 * time.Hour)

	task := tm.AddTask(title, description, priority, dueDate)

	assert.NotNil(t, task)
	assert.Equal(t, title, task.Title)
	assert.Equal(t, description, task.Description)
	assert.Equal(t, priority, task.Priority)
	assert.Equal(t, dueDate.Format("2006-01-02"), task.DueDate.Format("2006-01-02"))
	assert.False(t, task.Completed)
	assert.Equal(t, 1, task.ID)

	// Проверяем, что задача добавлена в список
	assert.Equal(t, 1, len(tm.tasks))
	assert.Equal(t, 2, tm.nextID)
}

func TestGetTask(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	task := tm.AddTask("Task 1", "Description", 1, time.Now())
	tm.AddTask("Task 2", "Description", 2, time.Now())

	foundTask := tm.GetTask(task.ID)
	assert.NotNil(t, foundTask)
	assert.Equal(t, task.ID, foundTask.ID)
	assert.Equal(t, task.Title, foundTask.Title)

	// Проверяем отсутствующую задачу
	notFoundTask := tm.GetTask(999)
	assert.Nil(t, notFoundTask)
}

func TestDeleteTask(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	task := tm.AddTask("Task to delete", "Description", 1, time.Now())

	// Удаляем существующую задачу
	success := tm.DeleteTask(task.ID)
	assert.True(t, success)
	assert.Equal(t, 0, len(tm.tasks))

	// Пытаемся удалить несуществующую задачу
	success = tm.DeleteTask(999)
	assert.False(t, success)
}

func TestUpdateTask(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	task := tm.AddTask("Original Title", "Original Description", 1, time.Now())

	newTitle := "Updated Title"
	newDescription := "Updated Description"
	newPriority := 3
	newDueDate := time.Now().Add(48 * time.Hour)
	newCompleted := true

	success := tm.UpdateTask(task.ID, newTitle, newDescription, newPriority, newDueDate, newCompleted)
	assert.True(t, success)

	updatedTask := tm.GetTask(task.ID)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, newTitle, updatedTask.Title)
	assert.Equal(t, newDescription, updatedTask.Description)
	assert.Equal(t, newPriority, updatedTask.Priority)
	assert.Equal(t, newDueDate.Format("2006-01-02"), updatedTask.DueDate.Format("2006-01-02"))
	assert.Equal(t, newCompleted, updatedTask.Completed)

	// Пытаемся обновить несуществующую задачу
	success = tm.UpdateTask(999, "Title", "Description", 1, time.Now(), false)
	assert.False(t, success)
}

func TestToggleTaskCompletion(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	task := tm.AddTask("Task to toggle", "Description", 2, time.Now())
	assert.False(t, task.Completed)

	// Переключаем статус
	success := tm.ToggleTaskCompletion(task.ID)
	assert.True(t, success)
	assert.True(t, tm.GetTask(task.ID).Completed)

	// Переключаем еще раз
	success = tm.ToggleTaskCompletion(task.ID)
	assert.True(t, success)
	assert.False(t, tm.GetTask(task.ID).Completed)

	// Пытаемся переключить несуществующую задачу
	success = tm.ToggleTaskCompletion(999)
	assert.False(t, success)
}

func TestSearchTasks(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	tm.AddTask("Important Meeting", "Discuss quarterly results", 3, time.Now())
	tm.AddTask("Buy Groceries", "Milk, bread, eggs", 2, time.Now())
	tm.AddTask("Call Mom", "Wish her happy birthday", 1, time.Now())

	// Поиск по заголовку
	results := tm.SearchTasks("Meeting")
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Important Meeting", results[0].Title)

	// Поиск по описанию
	results = tm.SearchTasks("birthday")
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Call Mom", results[0].Title)

	// Поиск без результатов
	results = tm.SearchTasks("nothing")
	assert.Equal(t, 0, len(results))

	// Поиск с несколькими результатами
	results = tm.SearchTasks("e")
	assert.True(t, len(results) >= 2) // Должно быть минимум 2 совпадения
}

func TestFilterTasksByStatus(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	// Создаем задачи с разными статусами
	tm.AddTask("Task 1", "Description", 1, time.Now())
	t2 := tm.AddTask("Task 2", "Description", 2, time.Now())
	tm.AddTask("Task 3", "Description", 3, time.Now())

	// Помечаем вторую задачу как выполненную
	tm.ToggleTaskCompletion(t2.ID)

	// Фильтруем активные задачи (не выполненные)
	activeTasks := tm.FilterTasksByStatus(false)
	assert.Equal(t, 2, len(activeTasks), "Должно быть 2 активные задачи")

	// Проверяем, что все активные задачи не помечены как выполненные
	for _, task := range activeTasks {
		assert.False(t, task.Completed, "Активная задача не должна быть помечена как выполненная")
	}

	// Фильтруем выполненные задачи
	completedTasks := tm.FilterTasksByStatus(true)
	assert.Equal(t, 1, len(completedTasks), "Должна быть 1 выполненная задача")

	// Проверяем, что выполненная задача помечена как выполненная
	assert.True(t, completedTasks[0].Completed, "Задача должна быть помечена как выполненная")
	assert.Equal(t, t2.ID, completedTasks[0].ID, "ID выполненной задачи должен соответствовать")
}

func TestSortTasksByPriority(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	tm.AddTask("Low priority", "Description", 1, time.Now())
	tm.AddTask("High priority", "Description", 3, time.Now())
	tm.AddTask("Medium priority", "Description", 2, time.Now())

	// Сортируем по приоритету
	sortedTasks := tm.SortTasksByPriority()

	// Проверяем порядок: сначала высокий приоритет, затем средний, затем низкий
	assert.Equal(t, 3, sortedTasks[0].Priority)
	assert.Equal(t, "High priority", sortedTasks[0].Title)

	assert.Equal(t, 2, sortedTasks[1].Priority)
	assert.Equal(t, "Medium priority", sortedTasks[1].Title)

	assert.Equal(t, 1, sortedTasks[2].Priority)
	assert.Equal(t, "Low priority", sortedTasks[2].Title)
}

func TestSaveAndLoadFromFile(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	// Создаем несколько задач
	tm.AddTask("Task 1", "Description 1", 1, time.Now())
	tm.AddTask("Task 2", "Description 2", 2, time.Now().Add(24*time.Hour))
	tm.AddTask("Task 3", "Description 3", 3, time.Now().Add(48*time.Hour))

	// Сохраняем в файл
	err := tm.SaveToFile()
	assert.NoError(t, err)

	// Проверяем, что файл создан
	_, err = os.Stat(testFilename)
	assert.False(t, os.IsNotExist(err))

	// Создаем новый менеджер и загружаем данные
	tm2 := NewTaskManager(testFilename)
	err = tm2.LoadFromFile()
	assert.NoError(t, err)

	// Проверяем загруженные данные
	assert.Equal(t, 3, len(tm2.tasks))
	assert.Equal(t, 4, tm2.nextID) // nextID должен быть равен последнему ID + 1

	// Проверяем содержимое задач
	assert.Equal(t, "Task 1", tm2.tasks[0].Title)
	assert.Equal(t, "Task 2", tm2.tasks[1].Title)
	assert.Equal(t, "Task 3", tm2.tasks[2].Title)

	// Проверяем приоритеты и другие поля
	assert.Equal(t, 1, tm2.tasks[0].Priority)
	assert.Equal(t, 2, tm2.tasks[1].Priority)
	assert.Equal(t, 3, tm2.tasks[2].Priority)
}

func TestExportToCSV(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	// Создаем задачи для экспорта
	t1 := tm.AddTask("Task 1", "Description 1", 1, time.Now())
	tm.AddTask("Task 2", "Description 2", 3, time.Now().Add(24*time.Hour))

	// Помечаем первую задачу как выполненную
	tm.ToggleTaskCompletion(t1.ID)

	// Экспортируем в CSV
	err := tm.ExportToCSV(testCSVFilename)
	assert.NoError(t, err, "Экспорт в CSV не должен вызывать ошибок")

	// Проверяем, что файл создан
	_, err = os.Stat(testCSVFilename)
	assert.False(t, os.IsNotExist(err), "Файл CSV должен быть создан")

	// Считываем содержимое файла
	file, err := os.Open(testCSVFilename)
	assert.NoError(t, err, "Не удалось открыть файл CSV")
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	assert.NoError(t, err, "Не удалось прочитать содержимое CSV файла")

	// Проверяем количество записей (заголовок + 2 задачи)
	assert.Equal(t, 3, len(records), "В CSV файле должно быть 3 записи (заголовок + 2 задачи)")

	// Проверяем заголовки
	assert.Equal(t, []string{"ID", "Title", "Description", "Priority", "Due Date", "Created At", "Completed"}, records[0])

	// Проверяем первую задачу
	assert.Contains(t, records[1][1], "Task 1", "Первая задача должна содержать 'Task 1'")
	assert.Contains(t, records[1][3], "Low", "Первая задача должна иметь приоритет 'Low'")
	assert.Contains(t, records[1][6], "Yes", "Первая задача должна быть помечена как выполненная (Yes)")

	// Проверяем вторую задачу
	assert.Contains(t, records[2][1], "Task 2", "Вторая задача должна содержать 'Task 2'")
	assert.Contains(t, records[2][3], "High", "Вторая задача должна иметь приоритет 'High'")
	assert.Contains(t, records[2][6], "No", "Вторая задача должна быть помечена как невыполненная (No)")
}

func TestSortTasksByDueDate(t *testing.T) {
	defer teardownTestManager()
	tm := setupTestManager()

	// Создаем задачи с разными сроками выполнения
	now := time.Now()
	t1 := tm.AddTask("Task 1", "Due tomorrow", 2, now.Add(24*time.Hour))
	t2 := tm.AddTask("Task 2", "Due today", 3, now) // Сегодня
	t3 := tm.AddTask("Task 3", "Due in a week", 1, now.Add(7*24*time.Hour))

	// Сортируем по сроку выполнения
	sortedTasks := tm.SortTasksByDueDate()

	// Проверяем порядок: сначала сегодня, потом завтра, потом через неделю
	assert.Equal(t, t2.ID, sortedTasks[0].ID) // Сегодня
	assert.Equal(t, t1.ID, sortedTasks[1].ID) // Завтра
	assert.Equal(t, t3.ID, sortedTasks[2].ID) // Через неделю

	// Проверяем, что даты в правильном порядке
	assert.True(t, sortedTasks[0].DueDate.Before(sortedTasks[1].DueDate))
	assert.True(t, sortedTasks[1].DueDate.Before(sortedTasks[2].DueDate))
}
