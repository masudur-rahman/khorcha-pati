package repos

import "github.com/masudur-rahman/khorcha-pati/models"

type EventRepository interface {
	AddEvent(event string) error
	ListEvents() ([]models.Event, error)
}
