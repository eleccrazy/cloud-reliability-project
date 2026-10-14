package service

import "errors"

var ErrItemNotFound = errors.New("item not found")

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type ItemService struct {
	items []Item
}

func NewItemService() *ItemService {
	return &ItemService{
		items: []Item{
			{ID: 1, Name: "Keyboard", Price: 80},
			{ID: 2, Name: "Mouse", Price: 40},
			{ID: 3, Name: "Monitor", Price: 250},
		},
	}
}

func (s *ItemService) List() []Item {
	return s.items
}

func (s *ItemService) GetByID(id int) (Item, error) {
	for _, item := range s.items {
		if item.ID == id {
			return item, nil
		}
	}

	return Item{}, ErrItemNotFound
}
