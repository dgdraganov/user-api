package repository

type User struct {
	ID           string `gorm:"type:varchar(36);primaryKey;autoIncrement:false"`
	FirstName    string `gorm:"type:varchar(255);not null"`
	LastName     string `gorm:"type:varchar(255);not null"`
	PasswordHash string `gorm:"not null"`
	Age          int    `gorm:"not null;check:age >= 18"`
	Email        string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Files        []File `gorm:"foreignKey:UserID;references:ID"`
}

type File struct {
	ID      string `gorm:"primaryKey;autoIncrement:false"`
	UserID  string `gorm:"type:varchar(36);not null"`
	Content []byte `gorm:"type:mediumblob"`
	User    User   `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
