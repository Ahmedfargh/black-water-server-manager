package Managers

import (
	"fmt"

	BackgroundTask "github.com/ahmedfargh/server-manager/BackGround"
)

type BackgroundTaskManager struct {
	Tasks []BackgroundTask.BackgroundTask
}

func (m *BackgroundTaskManager) AddTask(task BackgroundTask.BackgroundTask) {
	m.Tasks = append(m.Tasks, task)
}

func (m *BackgroundTaskManager) RunAllTasks() {
	for _, task := range m.Tasks {
		fmt.Println("Task Run")
		go func(t BackgroundTask.BackgroundTask) {
			if _, err := t.Run(); err != nil {
				fmt.Println(err)
				t.HandleError(err)
			}
		}(task)
	}
}
