export interface Task {
  id: number
  user_id: number
  title: string
  description: string
  status_id: number
  priority: 'low' | 'medium' | 'high'
  due_date: string | null
  created_at: string
  updated_at: string
}

export interface CreateTaskBody {
  title: string
  description?: string
  status_id?: number
  priority?: 'low' | 'medium' | 'high'
  due_date?: string
}

export interface UpdateTaskBody {
  title: string
  description?: string
  status_id?: number
  priority?: 'low' | 'medium' | 'high'
  due_date?: string
}
