package services

import "github.com/masudur-rahman/khorcha-pati/models"

type EventService interface {
	AddEvent(event string) error
	ListEvents() ([]models.Event, error)
}
