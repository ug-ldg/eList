package model

type Stats struct {
	TotalTasks int `json:"total_tasks"`
	Pending    int `json:"pending"`
	Done       int `json:"done"`
	RootTasks  int `json:"root_tasks"`
	SubTasks   int `json:"sub_tasks"`
}
