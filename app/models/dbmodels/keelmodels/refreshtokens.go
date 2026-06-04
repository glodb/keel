package keelmodels

type RefreshToken struct {
	UserId       int    `db:"userId INTEGER UNIQUE NOT NULL"`
	RefreshToken string `db:"refreshToken VARCHAR(255)" json:"refreshToken"`
}
