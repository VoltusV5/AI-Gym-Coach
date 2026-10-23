package users_postgres_repository

import (
	"context"
	"errors"
	"testing"
	"time"

	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetListNotes проверяет завершение чтения страницы, включая ошибку после частичного результата.
func TestGetListNotes(t *testing.T) {
	iterationErr := errors.New("connection interrupted")
	for _, tt := range []struct {
		name string
		ids  []int
		err  error
	}{
		{name: "empty page"},
		{name: "complete page", ids: []int{7, 3}},
		{name: "error before first row", err: iterationErr},
		{name: "error after partial page", ids: []int{7}, err: iterationErr},
		{name: "context canceled", ids: []int{7}, err: context.Canceled},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rows := &noteRows{ids: tt.ids, err: tt.err}
			repo := NewUsersRepository(&notesPool{rows: rows})

			notes, total, err := repo.GetListNotes(context.Background(), "10", 20, 0)

			assert.True(t, rows.closed, "rows must be closed on success and failure")
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				assert.Empty(t, notes, "a failed query must not return a partial page")
				assert.Zero(t, total)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, len(tt.ids), total)
			require.NotNil(t, notes, "an empty page must remain a JSON array")
			require.Len(t, notes, len(tt.ids))
			for i, id := range tt.ids {
				assert.Equal(t, id, notes[i].ID)
			}
		})
	}
}

type notesPool struct {
	core_postgres_pool.Pool
	rows *noteRows
}

// OpTimeout задаёт таймаут тестового репозитория без подключения к БД.
func (p *notesPool) OpTimeout() time.Duration { return time.Second }

// QueryRow возвращает число строк тестовой страницы для COUNT(*).
func (p *notesPool) QueryRow(context.Context, string, ...any) core_postgres_pool.Row {
	return notesCount(len(p.rows.ids))
}

// Query возвращает управляемый итератор вместо реального SQL-запроса.
func (p *notesPool) Query(context.Context, string, ...any) (core_postgres_pool.Rows, error) {
	return p.rows, nil
}

type notesCount int

// Scan записывает число заметок в результат COUNT(*).
func (n notesCount) Scan(dest ...any) error {
	*dest[0].(*int) = int(n)
	return nil
}

type noteRows struct {
	ids    []int
	index  int
	err    error
	closed bool
}

// Close фиксирует освобождение итератора репозиторием.
func (r *noteRows) Close() { r.closed = true }

// Err возвращает ошибку потока, возникающую после чтения доступных строк.
func (r *noteRows) Err() error { return r.err }

// Next переводит итератор на следующую доступную заметку.
func (r *noteRows) Next() bool {
	if r.index >= len(r.ids) {
		return false
	}
	r.index++
	return true
}

// Scan заполняет ID текущей заметки; остальные поля в этих сценариях несущественны.
func (r *noteRows) Scan(dest ...any) error {
	*dest[0].(*int) = r.ids[r.index-1]
	return nil
}
