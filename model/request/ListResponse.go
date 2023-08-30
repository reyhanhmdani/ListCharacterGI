package request

import "ListCharacterGI/model/entity"

type ListResponse struct {
	Status  interface{}       `json:"status"`
	Message interface{}       `json:"message"`
	Data    entity.Characters `json:"data"`
}

type ResponseToGetAll struct {
	Message string              `json:"message"`
	Data    int                 `json:"data"`
	MHS     []entity.Characters `json:"todos"`
}

type IDResponse struct {
	Message interface{} `json:"message"`
	Data    interface{} `json:"data"`
}

type DeleteResponse struct {
	Status  int         `json:"status"`
	Message interface{} `json:"message"`
}

type UpdateResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"data"`
	MHS     interface{} `json:"todos"`
}
