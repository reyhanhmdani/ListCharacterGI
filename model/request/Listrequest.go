package request

import "ListCharacterGI/model/entity"

type ListCreateRequest struct {
	Name       string            `json:"name" binding:"required"`
	Age        int               `json:"age"`
	Address    entity.Address    `json:"address"`
	Element    entity.Elements   `json:"element"`
	WeaponType entity.WeaponType `json:"weapon_type"`
	StarRating entity.StarRating `json:"star_rating"`
	UserID     int64             `json:"user_id"`
}

type ListUpdateRequest struct {
	Name       string            `json:"name"`
	Age        int               `json:"age"`
	Address    entity.Address    `json:"address"`
	Element    entity.Elements   `gorm:"type:varchar(255)" json:"element"`
	WeaponType entity.WeaponType `json:"weapon_type"`
	StarRating entity.StarRating `json:"star_rating"`
}
