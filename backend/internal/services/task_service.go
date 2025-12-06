package services

type TaskService interface {
    CreateTask(...) (*models.Task, error)
    GetTaskByID(...) (*models.Task, error)
    // ...
}
