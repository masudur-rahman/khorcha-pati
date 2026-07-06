package event

import (
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/repos"
)

type eventService struct {
	er repos.EventRepository
}

func NewEventService(er repos.EventRepository) *eventService {
	return &eventService{er: er}
}

func (e *eventService) AddEvent(event string) error {
	//TODO implement me
	panic("implement me")
}

func (e *eventService) ListEvents() ([]models.Event, error) {
	//TODO implement me
	panic("implement me")
}
