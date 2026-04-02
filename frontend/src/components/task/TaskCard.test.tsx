import type { Task } from '@/types/task'
import { render, screen } from '@testing-library/react'
import { type ReactElement } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { TaskCard } from './TaskCard'

const mockTask: Task = {
  id: 1,
  user_id: 1,
  title: 'テストタスク',
  description: 'テストの説明',
  status_id: 1,
  priority: 'medium',
  due_date: null,
  created_at: '2026-04-01T00:00:00+09:00',
  updated_at: '2026-04-01T00:00:00+09:00',
}

const renderWithRouter = (ui: ReactElement) => render(<MemoryRouter>{ui}</MemoryRouter>)

test('タイトルが表示される', () => {
  renderWithRouter(<TaskCard task={mockTask} />)
  expect(screen.getByText('テストタスク')).toBeInTheDocument()
})

test('完了済みタスクは半透明で表示される', () => {
  const doneTask: Task = { ...mockTask, status_id: 3 }
  renderWithRouter(<TaskCard task={doneTask} />)
  expect(screen.getByRole('article')).toHaveClass('opacity-50')
})

test('優先度ラベルが表示される', () => {
  renderWithRouter(<TaskCard task={mockTask} />)
  expect(screen.getByText('中')).toBeInTheDocument()
})
