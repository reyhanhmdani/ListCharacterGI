package entity

type Characters struct {
	ID          int64        `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(255);uniqueIndex:idx_email_name" json:"name"`
	Age         int          `json:"age"`
	Address     Address      `json:"address"`
	Element     Elements     `json:"element"`
	WeaponType  WeaponType   `json:"weapon_type"`
	UserID      int64        `json:"user_id"`
	StarRating  StarRating   `json:"star_rating"`
	Attachments []Attachment `gorm:"foreignKey:user_id" json:"attachments"`
}

type Address string

const (
	Mondstadt Elements = "Mondstadt"
	Liyue     Elements = "Liyue"
	Inazuma   Elements = "Inazuma"
	Sumeru    Elements = "Sumeru"
	Fontaine  Elements = "Fontaine"
	Natlan    Elements = "Natlan"
	Snezhnaya Elements = "Snezhnaya"
)

type Elements string

const (
	ElementsFire    Elements = "Pyro"
	ElementsCyro    Elements = "Cyro"
	ElementsAnemo   Elements = "Anemo"
	ElementsGeo     Elements = "Geo"
	ElementsElectro Elements = "Electro"
	ElementsHydro   Elements = "Hydro"
	ElementsDendro  Elements = "Dendro"
)

type WeaponType string

const (
	Claimore WeaponType = "Claimore"
	Sword    WeaponType = "Sword"
	Polearm  WeaponType = "Polearm"
	Catalyst WeaponType = "Catalyst"
	Bow      WeaponType = "Bow"
)

type StarRating string

const (
	StarRating4 StarRating = "4"
	StarRating5 StarRating = "5"
)
