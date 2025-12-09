package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Task представляет одну задачу
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"` // 1 - низкий, 2 - средний, 3 - высокий
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	Completed   bool      `json:"completed"`
}

// TaskManager управляет списком задач
type TaskManager struct {
	tasks    []*Task
	nextID   int
	filename string
}

// NewTaskManager создает новый менеджер задач
func NewTaskManager(filename string) *TaskManager {
	return &TaskManager{
		tasks:    []*Task{},
		nextID:   1,
		filename: filename,
	}
}

// AddTask добавляет новую задачу
func (tm *TaskManager) AddTask(title, description string, priority int, dueDate time.Time) *Task {
	task := &Task{
		ID:          tm.nextID,
		Title:       title,
		Description: description,
		Priority:    priority,
		DueDate:     dueDate,
		CreatedAt:   time.Now(),
		Completed:   false,
	}

	tm.tasks = append(tm.tasks, task)
	tm.nextID++
	return task
}

// GetTask возвращает задачу по ID
func (tm *TaskManager) GetTask(id int) *Task {
	for _, task := range tm.tasks {
		if task.ID == id {
			return task
		}
	}
	return nil
}

// DeleteTask удаляет задачу по ID
func (tm *TaskManager) DeleteTask(id int) bool {
	for i, task := range tm.tasks {
		if task.ID == id {
			tm.tasks = append(tm.tasks[:i], tm.tasks[i+1:]...)
			return true
		}
	}
	return false
}

// UpdateTask обновляет существующую задачу
func (tm *TaskManager) UpdateTask(id int, title, description string, priority int, dueDate time.Time, completed bool) bool {
	task := tm.GetTask(id)
	if task == nil {
		return false
	}

	task.Title = title
	task.Description = description
	task.Priority = priority
	task.DueDate = dueDate
	task.Completed = completed
	return true
}

// ToggleTaskCompletion изменяет статус выполнения задачи
func (tm *TaskManager) ToggleTaskCompletion(id int) bool {
	task := tm.GetTask(id)
	if task == nil {
		return false
	}

	task.Completed = !task.Completed
	return true
}

// SearchTasks ищет задачи по ключевому слову
func (tm *TaskManager) SearchTasks(keyword string) []*Task {
	keyword = strings.ToLower(keyword)
	var results []*Task

	for _, task := range tm.tasks {
		if strings.Contains(strings.ToLower(task.Title), keyword) ||
			strings.Contains(strings.ToLower(task.Description), keyword) {
			results = append(results, task)
		}
	}

	return results
}

// FilterTasksByStatus фильтрует задачи по статусу
func (tm *TaskManager) FilterTasksByStatus(completed bool) []*Task {
	var results []*Task

	for _, task := range tm.tasks {
		if task.Completed == completed {
			results = append(results, task)
		}
	}

	return results
}

// SortTasksByPriority сортирует задачи по приоритету
func (tm *TaskManager) SortTasksByPriority() []*Task {
	sortedTasks := make([]*Task, len(tm.tasks))
	copy(sortedTasks, tm.tasks)

	sort.Slice(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].Priority > sortedTasks[j].Priority
	})

	return sortedTasks
}

// SortTasksByDueDate сортирует задачи по сроку выполнения
func (tm *TaskManager) SortTasksByDueDate() []*Task {
	sortedTasks := make([]*Task, len(tm.tasks))
	copy(sortedTasks, tm.tasks)

	sort.Slice(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].DueDate.Before(sortedTasks[j].DueDate)
	})

	return sortedTasks
}

// SaveToFile сохраняет задачи в файл
func (tm *TaskManager) SaveToFile() error {
	data, err := json.MarshalIndent(tm.tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tm.filename, data, 0644)
}

// LoadFromFile загружает задачи из файла
func (tm *TaskManager) LoadFromFile() error {
	data, err := os.ReadFile(tm.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Файл не существует, это нормально для первого запуска
		}
		return err
	}

	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return err
	}

	tm.tasks = tasks

	// Обновляем nextID
	for _, task := range tm.tasks {
		if task.ID >= tm.nextID {
			tm.nextID = task.ID + 1
		}
	}

	return nil
}

// ExportToCSV экспортирует задачи в CSV формат
func (tm *TaskManager) ExportToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записываем заголовки
	headers := []string{"ID", "Title", "Description", "Priority", "Due Date", "Created At", "Completed"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Записываем данные
	for _, task := range tm.tasks {
		priorityText := map[int]string{1: "Low", 2: "Medium", 3: "High"}[task.Priority]
		completedText := "No"
		if task.Completed {
			completedText = "Yes"
		}

		// Используем правильный формат даты как в тестах
		row := []string{
			strconv.Itoa(task.ID),
			task.Title,
			task.Description,
			priorityText,
			task.DueDate.Format("2006-01-02 15:04"),
			task.CreatedAt.Format("2006-01-02 15:04"),
			completedText,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// Вспомогательные функции для диалоговых окон

func showAddTaskDialog(w fyne.Window, tm *TaskManager, updateList func()) {
	titleEntry := widget.NewEntry()
	descEntry := widget.NewMultiLineEntry()
	prioritySelect := widget.NewSelect([]string{"Low (1)", "Medium (2)", "High (3)"}, nil)
	prioritySelect.SetSelected("Medium (2)")

	// Устанавливаем сегодняшнюю дату как значение по умолчанию
	now := time.Now()
	dueDateEntry := widget.NewEntry()
	dueDateEntry.SetText(now.Add(24 * time.Hour).Format("2006-01-02"))

	formItems := []*widget.FormItem{
		{Text: "Title", Widget: titleEntry},
		{Text: "Description", Widget: descEntry},
		{Text: "Priority", Widget: prioritySelect},
		{Text: "Due Date (YYYY-MM-DD)", Widget: dueDateEntry},
	}

	dialog.ShowForm("Add New Task", "Add", "Cancel", formItems, func(confirmed bool) {
		if confirmed {
			// Парсим приоритет
			priority := 2
			switch prioritySelect.Selected {
			case "Low (1)":
				priority = 1
			case "Medium (2)":
				priority = 2
			case "High (3)":
				priority = 3
			}

			// Парсим дату
			dueDate, err := time.Parse("2006-01-02", dueDateEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid date format, use YYYY-MM-DD"), w)
				return
			}

			// Добавляем задачу
			tm.AddTask(titleEntry.Text, descEntry.Text, priority, dueDate)
			updateList()
		}
	}, w)
}

func showEditTaskDialog(w fyne.Window, tm *TaskManager, task *Task, updateList func()) {
	titleEntry := widget.NewEntry()
	titleEntry.SetText(task.Title)

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetText(task.Description)

	prioritySelect := widget.NewSelect([]string{"Low (1)", "Medium (2)", "High (3)"}, nil)
	switch task.Priority {
	case 1:
		prioritySelect.SetSelected("Low (1)")
	case 2:
		prioritySelect.SetSelected("Medium (2)")
	case 3:
		prioritySelect.SetSelected("High (3)")
	}

	dueDateEntry := widget.NewEntry()
	dueDateEntry.SetText(task.DueDate.Format("2006-01-02"))

	completedCheck := widget.NewCheck("Completed", nil)
	completedCheck.SetChecked(task.Completed)

	formItems := []*widget.FormItem{
		{Text: "Title", Widget: titleEntry},
		{Text: "Description", Widget: descEntry},
		{Text: "Priority", Widget: prioritySelect},
		{Text: "Due Date (YYYY-MM-DD)", Widget: dueDateEntry},
		{Text: "Status", Widget: completedCheck},
	}

	dialog.ShowForm("Edit Task", "Save", "Cancel", formItems, func(confirmed bool) {
		if confirmed {
			// Парсим приоритет
			priority := 2
			switch prioritySelect.Selected {
			case "Low (1)":
				priority = 1
			case "Medium (2)":
				priority = 2
			case "High (3)":
				priority = 3
			}

			// Парсим дату
			dueDate, err := time.Parse("2006-01-02", dueDateEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid date format, use YYYY-MM-DD"), w)
				return
			}

			// Обновляем задачу
			tm.UpdateTask(task.ID, titleEntry.Text, descEntry.Text, priority, dueDate, completedCheck.Checked)
			updateList()
		}
	}, w)
}

// Основная функция приложения
func main() {
	a := app.New()
	w := a.NewWindow("Task Manager")
	w.Resize(fyne.NewSize(800, 600))

	tm := NewTaskManager("tasks.json")
	tm.LoadFromFile()

	// Данные для привязки к интерфейсу
	taskList := binding.NewStringList()
	selectedTaskID := binding.NewInt()

	// Обновляем список задач в интерфейсе
	updateTaskList := func() {
		var ids []string
		for _, task := range tm.tasks {
			status := " "
			if task.Completed {
				status = "✓"
			}
			priority := map[int]string{1: "низкий", 2: "средний", 3: "высокий"}[task.Priority]
			ids = append(ids, fmt.Sprintf("[%s] %s (приоритет: %s, до: %s)",
				status, task.Title, priority, task.DueDate.Format("2006-01-02")))
		}
		taskList.Set(ids)
	}

	// Инициализируем список
	updateTaskList()

	// Создаем интерфейс
	taskListView := widget.NewListWithData(
		taskList,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(data binding.DataItem, item fyne.CanvasObject) {
			item.(*widget.Label).Bind(data.(binding.String))
		},
	)

	// Обработка выбора задачи
	taskListView.OnSelected = func(id widget.ListItemID) {
		if id < len(tm.tasks) {
			selectedTaskID.Set(tm.tasks[id].ID)
		}
	}

	// Кнопки управления
	addButton := widget.NewButton("Добавить задачу", func() {
		showAddTaskDialog(w, tm, updateTaskList)
	})

	editButton := widget.NewButton("Редактировать", func() {
		id, _ := selectedTaskID.Get()
		task := tm.GetTask(id)
		if task != nil {
			showEditTaskDialog(w, tm, task, updateTaskList)
		} else {
			dialog.ShowInformation("Ошибка", "Выберите задачу для редактирования", w)
		}
	})

	deleteButton := widget.NewButton("Удалить", func() {
		id, _ := selectedTaskID.Get()
		if id > 0 {
			if tm.DeleteTask(id) {
				updateTaskList()
				selectedTaskID.Set(0)
			}
		}
	})

	toggleButton := widget.NewButton("Изменить статус", func() {
		id, _ := selectedTaskID.Get()
		if id > 0 {
			tm.ToggleTaskCompletion(id)
			updateTaskList()
		}
	})

	saveButton := widget.NewButton("Сохранить", func() {
		if err := tm.SaveToFile(); err == nil {
			dialog.ShowInformation("Успешно", "Задачи сохранены в файл", w)
		} else {
			dialog.ShowError(err, w)
		}
	})

	exportButton := widget.NewButton("Экспорт в CSV", func() {
		dialog.ShowFileSave(func(file fyne.URIWriteCloser, err error) {
			if file != nil {
				filename := file.URI().Path()
				file.Close()

				if err := tm.ExportToCSV(filename); err == nil {
					dialog.ShowInformation("Успешно", "Задачи экспортированы в CSV", w)
				} else {
					dialog.ShowError(err, w)
				}
			}
		}, w)
	})

	// Кнопка для сортировки по приоритету
	sortPriorityButton := widget.NewButton("Сортировка по приоритету", func() {
		tm.tasks = tm.SortTasksByPriority()
		updateTaskList()
	})

	// Кнопка для сортировки по дате выполнения
	sortDateButton := widget.NewButton("Сортировка по дате", func() {
		tm.tasks = tm.SortTasksByDueDate()
		updateTaskList()
	})

	// Поле для поиска
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск задач...")
	searchEntry.OnChanged = func(text string) {
		if text == "" {
			// Если поле поиска пустое, отображаем все задачи
			updateTaskList()
			return
		}

		// Ищем задачи по ключевому слову
		results := tm.SearchTasks(text)

		// Обновляем список отображаемых задач
		var ids []string
		for _, task := range results {
			status := " "
			if task.Completed {
				status = "✓"
			}
			priority := map[int]string{1: "низкий", 2: "средний", 3: "высокий"}[task.Priority]
			ids = append(ids, fmt.Sprintf("[%s] %s (приоритет: %s, до: %s)",
				status, task.Title, priority, task.DueDate.Format("2006-01-02")))
		}
		taskList.Set(ids)
	}

	// Чекбокс для фильтрации по статусу
	filterActive := widget.NewCheck("Показать только активные", func(checked bool) {
		if checked {
			// Показываем только активные (не выполненные) задачи
			filteredTasks := tm.FilterTasksByStatus(false)
			var ids []string
			for _, task := range filteredTasks {
				status := " "
				priority := map[int]string{1: "низкий", 2: "средний", 3: "высокий"}[task.Priority]
				ids = append(ids, fmt.Sprintf("[%s] %s (приоритет: %s, до: %s)",
					status, task.Title, priority, task.DueDate.Format("2006-01-02")))
			}
			taskList.Set(ids)
		} else {
			// Показываем все задачи
			updateTaskList()
		}
	})

	// Размещение элементов интерфейса
	buttonContainer := container.NewGridWithColumns(6, addButton, editButton, deleteButton, toggleButton, saveButton, exportButton)
	sortContainer := container.NewGridWithColumns(2, sortPriorityButton, sortDateButton)
	filterContainer := container.NewBorder(nil, nil, nil, nil, filterActive, searchEntry)

	mainContainer := container.NewVBox(
		filterContainer,
		widget.NewSeparator(),
		taskListView,
	)

	content := container.NewBorder(
		container.NewVBox(buttonContainer, sortContainer),
		nil, nil, nil,
		mainContainer,
	)

	w.SetContent(content)
	w.ShowAndRun()
}
