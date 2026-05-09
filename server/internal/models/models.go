package models

import (
	"time"

	"gorm.io/gorm"
)

// User ユーザー情報
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string         `gorm:"type:varchar(100);not null" json:"nickname"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Group 夫婦・家族などのグループ
type Group struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	OwnerID   uint           `gorm:"not null" json:"owner_id"` // グループ管理者（作成者）
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Members   []User         `gorm:"many2many:group_members;" json:"members"`
	Owner     User           `gorm:"foreignKey:OwnerID" json:"-"`
}

// Receipt レシート明細
type Receipt struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	GroupID        uint           `gorm:"not null" json:"group_id"`
	UserID         uint           `gorm:"not null" json:"user_id"` // 入力したユーザー
	Date           time.Time      `gorm:"not null" json:"date"`
	SettlementYear int            `gorm:"not null" json:"settlement_year"`
	SettlementMonth int           `gorm:"not null" json:"settlement_month"`
	Shop           string         `gorm:"type:varchar(255)" json:"shop"`
	Item           string         `gorm:"type:varchar(255)" json:"item"`
	Amount         int            `gorm:"not null" json:"amount"`
	PayerID        uint           `gorm:"not null" json:"payer_id"` // 実際に支払ったユーザー
	PaymentMethod  string         `gorm:"type:varchar(50);not null" json:"payment_method"` // "折半", "自分が10割", "全額相手負担" など
	SettledAt      *time.Time     `json:"settled_at"` // 精算済みの場合、その日時
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Group Group `gorm:"foreignKey:GroupID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
	Payer User  `gorm:"foreignKey:PayerID" json:"payer"`
}

// Payment Methods
const (
	PaymentMethodHalf  = "half"  // 折半
	PaymentMethodSelf  = "self"  // 自分が10割負担
	PaymentMethodOther = "other" // 全額相手負担
)

// Settlement 精算情報
type Settlement struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GroupID   uint      `gorm:"not null" json:"group_id"`
	Year      int       `gorm:"not null" json:"year"`
	Month     int       `gorm:"not null" json:"month"`
	Amount    int       `gorm:"not null" json:"amount"`
	SettledBy uint      `gorm:"not null" json:"settled_by"`
	CreatedAt time.Time `json:"created_at"`

	Group     Group     `gorm:"foreignKey:GroupID" json:"-"`
	SettledByUser User `gorm:"foreignKey:SettledBy" json:"settled_by_user"`
}
