package repository

type User struct {
	ID           string         `gorm:"type:varchar(36);primaryKey;autoIncrement:false"`
	FirstName    string         `gorm:"type:varchar(255);not null"`
	LastName     string         `gorm:"type:varchar(255);not null"`
	PasswordHash string         `gorm:"not null"`
	Age          int            `gorm:"not null;check:age >= 18"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Files        []FileMetadata `gorm:"foreignKey:UserID;references:ID"`
}

type FileMetadata struct {
	ID         string `gorm:"primaryKey;autoIncrement:false"`
	FileName   string `gorm:"type:varchar(255);not null"`
	BucketName string `gorm:"type:varchar(255);not null"`
	UserID     string `gorm:"type:varchar(36);not null"`
	User       User   `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
