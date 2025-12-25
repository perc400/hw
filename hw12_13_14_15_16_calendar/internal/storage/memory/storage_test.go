package memorystorage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"                                        //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	t.Run("create_and_lists", func(t *testing.T) {
		moscowLoc, err := time.LoadLocation("Europe/Moscow")
		require.NoError(t, err)
		ctx := context.TODO()
		st := New()
		event := storage.Event{
			ID:       uuid.New().String(),
			Datetime: time.Date(2025, time.December, 22, 0, 0, 44, 0, moscowLoc),
			Duration: 7 * 24 * time.Hour,
		}
		err = st.Create(ctx, event)
		require.NoError(t, err)

		expectedListDay := []storage.Event{}
		actualListDay, err := st.ListDay(ctx, time.Date(2025, time.December, 30, 1, 33, 44, 0, moscowLoc))
		require.NoError(t, err)
		require.Equal(t, expectedListDay, actualListDay)

		expectedListWeek := []storage.Event{
			event,
		}
		actualListWeek, err := st.ListWeek(ctx, time.Date(2025, time.December, 20, 1, 33, 44, 0, moscowLoc))
		require.NoError(t, err)
		require.Equal(t, expectedListWeek, actualListWeek)
	})

	t.Run("update_delete_events", func(t *testing.T) {
		moscowLoc, err := time.LoadLocation("Europe/Moscow")
		require.NoError(t, err)
		ctx := context.TODO()
		st := New()
		event := storage.Event{
			ID:       uuid.New().String(),
			Datetime: time.Date(2025, time.December, 22, 0, 33, 44, 0, moscowLoc),
			Duration: 7 * 24 * time.Hour,
		}
		err = st.Create(ctx, event)
		require.NoError(t, err)

		updatedEvent := storage.Event{
			Datetime: time.Date(2025, time.December, 23, 0, 33, 44, 0, moscowLoc),
			Duration: 4 * 24 * time.Hour,
		}
		err = st.Update(ctx, event.ID, updatedEvent)
		require.NoError(t, err)

		expectedListDay := []storage.Event{
			{
				ID:       event.ID,
				Datetime: time.Date(2025, time.December, 23, 0, 33, 44, 0, moscowLoc),
				Duration: 4 * 24 * time.Hour,
			},
		}
		actualListDay, err := st.ListDay(ctx, time.Date(2025, time.December, 23, 5, 33, 44, 0, moscowLoc))
		require.NoError(t, err)
		require.Equal(t, expectedListDay, actualListDay)

		err = st.Delete(ctx, actualListDay[0].ID)
		require.NoError(t, err)
		newActualListDay, err := st.ListDay(ctx, time.Date(2025, time.December, 23, 5, 33, 44, 0, moscowLoc))
		require.NoError(t, err)
		require.Empty(t, newActualListDay)
	})

	t.Run("delete_nonexistent_event", func(t *testing.T) {
		moscowLoc, err := time.LoadLocation("Europe/Moscow")
		require.NoError(t, err)
		ctx := context.TODO()
		st := New()
		event := storage.Event{
			ID:       uuid.New().String(),
			Datetime: time.Date(2025, time.December, 22, 0, 33, 44, 0, moscowLoc),
			Duration: 7 * 24 * time.Hour,
		}
		err = st.Create(ctx, event)
		require.NoError(t, err)
		err = st.Delete(ctx, uuid.New().String())
		require.ErrorIs(t, err, storage.ErrEventNotFound)
	})

	t.Run("overlaps", func(t *testing.T) {
		moscowLoc, err := time.LoadLocation("Europe/Moscow")
		require.NoError(t, err)
		ctx := context.TODO()
		st := New()
		e1 := storage.Event{
			ID:       uuid.New().String(),
			Datetime: time.Date(2025, time.December, 22, 0, 33, 44, 0, moscowLoc),
			Duration: 7 * 24 * time.Hour,
		}
		err = st.Create(ctx, e1)
		require.NoError(t, err)

		e2 := storage.Event{
			ID:       uuid.New().String(),
			Datetime: time.Date(2025, time.December, 23, 0, 33, 44, 0, moscowLoc),
			Duration: 1 * 24 * time.Hour,
		}
		err = st.Create(ctx, e2)
		require.ErrorIs(t, err, storage.ErrDateBusy)

		e3 := storage.Event{
			ID:       uuid.New().String(),
			Datetime: time.Date(2025, time.December, 29, 1, 33, 44, 0, moscowLoc),
			Duration: 1 * 24 * time.Hour,
		}
		err = st.Create(ctx, e3)
		require.NoError(t, err)

		updatedS3 := storage.Event{
			Datetime: time.Date(2025, time.December, 24, 1, 33, 44, 0, moscowLoc),
			Duration: 1 * 24 * time.Hour,
		}
		err = st.Update(ctx, e3.ID, updatedS3)
		require.ErrorIs(t, err, storage.ErrDateBusy)
	})

	t.Run("concurrent_create", func(t *testing.T) {
		ctx := context.TODO()
		st := New()
		var wg sync.WaitGroup
		workers := 100
		start := time.Now()

		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				e := storage.Event{
					ID:       uuid.New().String(),
					UserID:   1,
					Datetime: start.Add(time.Duration(i) * time.Hour),
					Duration: time.Minute,
				}

				err := st.Create(ctx, e)
				require.NoError(t, err)
			}()
		}

		wg.Wait()

		events, err := st.ListWeek(context.TODO(), start)
		require.NoError(t, err)
		require.Len(t, events, workers)
	})
}
