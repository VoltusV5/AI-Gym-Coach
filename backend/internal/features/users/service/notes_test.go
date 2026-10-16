package users_service_test

import (
	"context"
	"errors"
	"testing"

	core_errors "sport_app/internal/core/errors"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersService_Notes_PaginationAndIDOR(t *testing.T) {
	t.Run("GetListNotes returns paginated notes and total count", func(t *testing.T) {
		repo := &mockUsersRepository{
			GetListNotesFunc: func(ctx context.Context, userID string, limit, offset int) ([]users_postgres_repository.Note, int, error) {
				assert.Equal(t, "user-42", userID)
				assert.Equal(t, 10, limit)
				assert.Equal(t, 20, offset)

				return []users_postgres_repository.Note{
					{ID: 21, UserID: 42, Title: "Note 21", Body: "Text 21"},
					{ID: 22, UserID: 42, Title: "Note 22", Body: "Text 22"},
				}, 55, nil
			},
		}

		svc, _ := newTestUsersService(repo)
		notes, total, err := svc.GetListNotes(context.Background(), "user-42", 10, 20)
		require.NoError(t, err)
		assert.Equal(t, 55, total)
		assert.Len(t, notes, 2)
		assert.Equal(t, 21, notes[0].ID)
	})

	t.Run("GetListNotes empty page returns empty slice and total", func(t *testing.T) {
		repo := &mockUsersRepository{
			GetListNotesFunc: func(ctx context.Context, userID string, limit, offset int) ([]users_postgres_repository.Note, int, error) {
				return []users_postgres_repository.Note{}, 55, nil
			},
		}

		svc, _ := newTestUsersService(repo)
		notes, total, err := svc.GetListNotes(context.Background(), "user-42", 10, 100)
		require.NoError(t, err)
		assert.Equal(t, 55, total)
		assert.Empty(t, notes)
	})

	t.Run("DeleteNotesUser succeeds for note owner", func(t *testing.T) {
		var deletedID int
		repo := &mockUsersRepository{
			DeleteNotesUserFunc: func(ctx context.Context, userID string, noteID int) error {
				assert.Equal(t, "user-42", userID)
				deletedID = noteID
				return nil
			},
		}

		svc, _ := newTestUsersService(repo)
		err := svc.DeleteNotesUser(context.Background(), "user-42", 99)
		require.NoError(t, err)
		assert.Equal(t, 99, deletedID)
	})

	t.Run("DeleteNotesUser returns ErrNotFound for other user's note (IDOR protection)", func(t *testing.T) {
		repo := &mockUsersRepository{
			DeleteNotesUserFunc: func(ctx context.Context, userID string, noteID int) error {
				// Simulating repository 0 rows affected when note belongs to another user
				return core_errors.ErrNotFound
			},
		}

		svc, _ := newTestUsersService(repo)
		err := svc.DeleteNotesUser(context.Background(), "attacker-user-id", 99)
		require.Error(t, err)
		assert.True(t, errors.Is(err, core_errors.ErrNotFound), "expected ErrNotFound when accessing another user's note")
	})

	t.Run("UpdateNotesUser returns ErrNotFound when note belongs to another user", func(t *testing.T) {
		repo := &mockUsersRepository{
			GetNoteByIDFunc: func(ctx context.Context, userID string, noteID int) (users_postgres_repository.Note, error) {
				// Repository rejects query when user_id does not match note owner
				return users_postgres_repository.Note{}, core_errors.ErrNotFound
			},
		}

		svc, _ := newTestUsersService(repo)
		_, err := svc.UpdateNotesUser(context.Background(), "attacker-user-id", 99, "Hacked", "Payload")
		require.Error(t, err)
		assert.True(t, errors.Is(err, core_errors.ErrNotFound))
	})
}
